package message

import "math"

// messageType is the type used to encode the frame's type.
type messageType uint16

// SessionIndex is the type for session indices.
type SessionIndex uint16

// headerSize is the type used to encode the frame's header size
type headerSize uint16

// HeaderMax is the maximum header length, in bytes.
const headerMax = math.MaxUint16

const (
	commandSyncCallType    messageType = 'C'<<8 | 'C'
	commandSyncSuccessType messageType = 'C'<<8 | 'S'
	commandSyncFailureType messageType = 'C'<<8 | 'F'
	commandSyncErrorType   messageType = 'C'<<8 | 'E'

	commandAsyncCallType    messageType = 'A'<<8 | 'C'
	commandAsyncSuccessType messageType = 'A'<<8 | 'S'
	commandAsyncFailureType messageType = 'A'<<8 | 'F'
	commandAsyncErrorType   messageType = 'A'<<8 | 'E'

	commandExecuteType messageType = 'C'<<8 | 'X'

	sessionNotificationType messageType = 'N'<<8 | 'O'

	sessionCreateType  messageType = 'S'<<8 | 'C'
	sessionDestroyType messageType = 'S'<<8 | 'D'
)
