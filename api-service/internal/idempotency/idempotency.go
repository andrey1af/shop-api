package idempotency

import (
	"bytes"
	"context"
	cryptorand "crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/redis/go-redis/v9"
)

const (
	maxRequestBodySize = 1 << 20
	claimPollInterval  = 25 * time.Millisecond
	claimLease         = time.Minute
	redisKeyPrefix     = "shop-api:idempotency:"
)

var claimScript = redis.NewScript(`
local stored_hash = redis.call('HGET', KEYS[1], 'request_hash')
if not stored_hash then
    local redis_time = redis.call('TIME')
    local now = redis_time[1] * 1000 + math.floor(redis_time[2] / 1000)
    redis.call('HSET', KEYS[1],
        'request_hash', ARGV[1],
        'owner_token', ARGV[2],
        'locked_until', now + tonumber(ARGV[3]))
    redis.call('PEXPIRE', KEYS[1], ARGV[4])
    return {'reserved'}
end

if stored_hash ~= ARGV[1] then
    return {'conflict'}
end

local status = redis.call('HGET', KEYS[1], 'status_code')
if status then
    return {
        'replay',
        status,
        redis.call('HGET', KEYS[1], 'response_headers') or '',
        redis.call('HGET', KEYS[1], 'response_body') or ''
    }
end

local redis_time = redis.call('TIME')
local now = redis_time[1] * 1000 + math.floor(redis_time[2] / 1000)
local locked_until = tonumber(redis.call('HGET', KEYS[1], 'locked_until') or '0')
if locked_until <= now then
    redis.call('HSET', KEYS[1],
        'owner_token', ARGV[2],
        'locked_until', now + tonumber(ARGV[3]))
    redis.call('PEXPIRE', KEYS[1], ARGV[4])
    return {'reserved'}
end

return {'wait'}
`)

var completeScript = redis.NewScript(`
local owner = redis.call('HGET', KEYS[1], 'owner_token')
if owner ~= ARGV[1] or redis.call('HEXISTS', KEYS[1], 'status_code') == 1 then
    return 0
end

redis.call('HSET', KEYS[1],
    'status_code', ARGV[2],
    'response_headers', ARGV[3],
    'response_body', ARGV[4])
redis.call('PEXPIRE', KEYS[1], ARGV[5])
return 1
`)

type Response struct {
	Status int
	Header http.Header
	Body   []byte
}

type Reservation interface {
	Complete(context.Context, Response) error
	Abort(context.Context)
}

type Decision struct {
	Reservation Reservation
	Replay      *Response
	Conflict    bool
}

type Backend interface {
	Begin(context.Context, string, string) (Decision, error)
}

type Store struct {
	client redis.Scripter
	ttl    time.Duration
}

func NewStore(client redis.Scripter, ttl time.Duration) *Store {
	return &Store{client: client, ttl: ttl}
}

func (store *Store) Begin(ctx context.Context, key, requestHash string) (Decision, error) {
	ownerToken, err := newOwnerToken()
	if err != nil {
		return Decision{}, err
	}

	redisKey := redisKeyPrefix + key
	for {
		result, err := claimScript.Run(
			ctx,
			store.client,
			[]string{redisKey},
			requestHash,
			ownerToken,
			claimLease.Milliseconds(),
			store.ttl.Milliseconds(),
		).Slice()
		if err != nil {
			return Decision{}, fmt.Errorf("claim Redis idempotency key: %w", err)
		}

		decision, wait, err := store.parseClaim(result, redisKey, ownerToken)
		if err != nil || !wait {
			return decision, err
		}

		timer := time.NewTimer(claimPollInterval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return Decision{}, ctx.Err()
		case <-timer.C:
		}
	}
}

func (store *Store) parseClaim(result []any, redisKey, ownerToken string) (Decision, bool, error) {
	if len(result) == 0 {
		return Decision{}, false, fmt.Errorf("claim Redis idempotency key: empty response")
	}

	action, err := redisString(result[0])
	if err != nil {
		return Decision{}, false, err
	}
	switch action {
	case "reserved":
		return Decision{Reservation: &redisReservation{
			client:     store.client,
			redisKey:   redisKey,
			ownerToken: ownerToken,
			ttl:        store.ttl,
		}}, false, nil
	case "conflict":
		return Decision{Conflict: true}, false, nil
	case "wait":
		return Decision{}, true, nil
	case "replay":
		if len(result) != 4 {
			return Decision{}, false, fmt.Errorf("claim Redis idempotency key: invalid replay response")
		}
		statusValue, err := redisString(result[1])
		if err != nil {
			return Decision{}, false, err
		}
		status, err := strconv.Atoi(statusValue)
		if err != nil {
			return Decision{}, false, fmt.Errorf("decode Redis idempotency status: %w", err)
		}
		rawHeaders, err := redisBytes(result[2])
		if err != nil {
			return Decision{}, false, err
		}
		headers := make(http.Header)
		if len(rawHeaders) > 0 {
			if err := json.Unmarshal(rawHeaders, &headers); err != nil {
				return Decision{}, false, fmt.Errorf("decode Redis idempotency headers: %w", err)
			}
		}
		body, err := redisBytes(result[3])
		if err != nil {
			return Decision{}, false, err
		}
		return Decision{Replay: &Response{Status: status, Header: headers, Body: body}}, false, nil
	default:
		return Decision{}, false, fmt.Errorf("claim Redis idempotency key: unknown action %q", action)
	}
}

type redisReservation struct {
	client     redis.Scripter
	redisKey   string
	ownerToken string
	ttl        time.Duration
}

func (reservation *redisReservation) Complete(ctx context.Context, response Response) error {
	rawHeaders, err := json.Marshal(response.Header)
	if err != nil {
		return fmt.Errorf("encode Redis idempotency response headers: %w", err)
	}

	stored, err := completeScript.Run(
		ctx,
		reservation.client,
		[]string{reservation.redisKey},
		reservation.ownerToken,
		response.Status,
		rawHeaders,
		response.Body,
		reservation.ttl.Milliseconds(),
	).Int64()
	if err != nil {
		return fmt.Errorf("store Redis idempotency response: %w", err)
	}
	if stored != 1 {
		return fmt.Errorf("store Redis idempotency response: reservation is no longer owned")
	}

	return nil
}

func (*redisReservation) Abort(context.Context) {}

func Require(backend Backend, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("Idempotency-Key")
		if strings.TrimSpace(key) == "" || utf8.RuneCountInString(key) > 255 {
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Idempotency-Key must contain between 1 and 255 characters")
			return
		}

		body, err := io.ReadAll(io.LimitReader(r.Body, maxRequestBodySize+1))
		if err != nil {
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request payload")
			return
		}
		if len(body) > maxRequestBodySize {
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Request body exceeds 1 MiB")
			return
		}
		r.Body = io.NopCloser(bytes.NewReader(body))

		decision, err := backend.Begin(r.Context(), key, requestFingerprint(r, body))
		if err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
			return
		}
		if decision.Conflict {
			writeError(w, http.StatusConflict, "IDEMPOTENCY_KEY_REUSED", "Idempotency key is already used with a different request")
			return
		}
		if decision.Replay != nil {
			writeResponse(w, *decision.Replay)
			return
		}

		buffer := newBufferedResponse()
		next(buffer, r)
		response := buffer.Response()
		if err := decision.Reservation.Complete(r.Context(), response); err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
			return
		}
		writeResponse(w, response)
	}
}

func requestFingerprint(r *http.Request, body []byte) string {
	hash := sha256.New()
	_, _ = hash.Write([]byte(r.Method))
	_, _ = hash.Write([]byte{'\n'})
	_, _ = hash.Write([]byte(r.URL.RequestURI()))
	_, _ = hash.Write([]byte{'\n'})
	_, _ = hash.Write(body)
	return hex.EncodeToString(hash.Sum(nil))
}

type bufferedResponse struct {
	header http.Header
	status int
	body   bytes.Buffer
}

func newBufferedResponse() *bufferedResponse {
	return &bufferedResponse{header: make(http.Header)}
}

func (response *bufferedResponse) Header() http.Header {
	return response.header
}

func (response *bufferedResponse) WriteHeader(status int) {
	if response.status == 0 {
		response.status = status
	}
}

func (response *bufferedResponse) Write(data []byte) (int, error) {
	if response.status == 0 {
		response.status = http.StatusOK
	}
	return response.body.Write(data)
}

func (response *bufferedResponse) Response() Response {
	status := response.status
	if status == 0 {
		status = http.StatusOK
	}
	return Response{Status: status, Header: response.header.Clone(), Body: bytes.Clone(response.body.Bytes())}
}

func writeResponse(w http.ResponseWriter, response Response) {
	for name, values := range response.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}
	w.WriteHeader(response.Status)
	_, _ = w.Write(response.Body)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"code": code, "message": message})
}

func newOwnerToken() (string, error) {
	raw := make([]byte, 16)
	if _, err := cryptorand.Read(raw); err != nil {
		return "", fmt.Errorf("generate idempotency owner token: %w", err)
	}
	return hex.EncodeToString(raw), nil
}

func redisString(value any) (string, error) {
	switch typed := value.(type) {
	case string:
		return typed, nil
	case []byte:
		return string(typed), nil
	default:
		return "", fmt.Errorf("decode Redis idempotency value: unexpected type %T", value)
	}
}

func redisBytes(value any) ([]byte, error) {
	switch typed := value.(type) {
	case string:
		return []byte(typed), nil
	case []byte:
		return bytes.Clone(typed), nil
	default:
		return nil, fmt.Errorf("decode Redis idempotency value: unexpected type %T", value)
	}
}
