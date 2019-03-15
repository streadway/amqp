// Copyright (c) 2012, Sean Treadway, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/streadway/amqp

/* GENERATED FILE - DO NOT EDIT */
/* Rebuild from the spec/gen.go tool */

package proto

import (
	"encoding/binary"
	"fmt"
	"io"
)

// Error codes that can be sent from the server during a connection or
// channel exception or used by the client to indicate a class of error like
// ErrCredentials.  The text of the error is likely more interesting than
// these constants.
const (
	FrameMethod        = 1
	FrameHeader        = 2
	FrameBody          = 3
	FrameHeartbeat     = 8
	FrameMinSize       = 4096
	FrameEnd           = 206
	ReplySuccess       = 200
	ContentTooLarge    = 311
	NoRoute            = 312
	NoConsumers        = 313
	ConnectionForced   = 320
	InvalidPath        = 402
	AccessRefused      = 403
	NotFound           = 404
	ResourceLocked     = 405
	PreconditionFailed = 406
	FrameError         = 501
	SyntaxError        = 502
	CommandInvalid     = 503
	ChannelError       = 504
	UnexpectedFrame    = 505
	ResourceError      = 506
	NotAllowed         = 530
	NotImplemented     = 540
	InternalError      = 541
)

// IsSoftExceptionCode returns true if the exception code can be recovered
func IsSoftExceptionCode(code int) bool {
	switch code {
	case 311:
		return true
	case 312:
		return true
	case 313:
		return true
	case 403:
		return true
	case 404:
		return true
	case 405:
		return true
	case 406:
		return true

	}
	return false
}

// ConnectionStart represents the AMQP message connection.start
type ConnectionStart struct {
	VersionMajor     byte
	VersionMinor     byte
	ServerProperties Table
	Mechanisms       string
	Locales          string
}

// ID returns the AMQP class and method identifiers for this message
func (msg *ConnectionStart) ID() (uint16, uint16) {
	return 10, 10
}

// Wait returns true when the client should expect a response from the server
func (msg *ConnectionStart) Wait() bool {
	return true
}

// Write serializes this message to the provided writer
func (msg *ConnectionStart) Write(w io.Writer) (err error) {

	if err = binary.Write(w, binary.BigEndian, msg.VersionMajor); err != nil {
		return
	}
	if err = binary.Write(w, binary.BigEndian, msg.VersionMinor); err != nil {
		return
	}

	if err = writeTable(w, msg.ServerProperties); err != nil {
		return
	}

	if err = writeLongstr(w, msg.Mechanisms); err != nil {
		return
	}
	if err = writeLongstr(w, msg.Locales); err != nil {
		return
	}

	return
}

// Read deserializes this message from the provided reader
func (msg *ConnectionStart) Read(r io.Reader) (err error) {

	if err = binary.Read(r, binary.BigEndian, &msg.VersionMajor); err != nil {
		return
	}
	if err = binary.Read(r, binary.BigEndian, &msg.VersionMinor); err != nil {
		return
	}

	if msg.ServerProperties, err = readTable(r); err != nil {
		return
	}

	if msg.Mechanisms, err = readLongstr(r); err != nil {
		return
	}
	if msg.Locales, err = readLongstr(r); err != nil {
		return
	}

	return
}

// ConnectionStartOk represents the AMQP message connection.start-ok
type ConnectionStartOk struct {
	ClientProperties Table
	Mechanism        string
	Response         string
	Locale           string
}

// ID returns the AMQP class and method identifiers for this message
func (msg *ConnectionStartOk) ID() (uint16, uint16) {
	return 10, 11
}

// Wait returns true when the client should expect a response from the server
func (msg *ConnectionStartOk) Wait() bool {
	return true
}

// Write serializes this message to the provided writer
func (msg *ConnectionStartOk) Write(w io.Writer) (err error) {

	if err = writeTable(w, msg.ClientProperties); err != nil {
		return
	}

	if err = writeShortstr(w, msg.Mechanism); err != nil {
		return
	}

	if err = writeLongstr(w, msg.Response); err != nil {
		return
	}

	if err = writeShortstr(w, msg.Locale); err != nil {
		return
	}

	return
}

// Read deserializes this message from the provided reader
func (msg *ConnectionStartOk) Read(r io.Reader) (err error) {

	if msg.ClientProperties, err = readTable(r); err != nil {
		return
	}

	if msg.Mechanism, err = readShortstr(r); err != nil {
		return
	}

	if msg.Response, err = readLongstr(r); err != nil {
		return
	}

	if msg.Locale, err = readShortstr(r); err != nil {
		return
	}

	return
}

// ConnectionSecure represents the AMQP message connection.secure
type ConnectionSecure struct {
	Challenge string
}

// ID returns the AMQP class and method identifiers for this message
func (msg *ConnectionSecure) ID() (uint16, uint16) {
	return 10, 20
}

// Wait returns true when the client should expect a response from the server
func (msg *ConnectionSecure) Wait() bool {
	return true
}

// Write serializes this message to the provided writer
func (msg *ConnectionSecure) Write(w io.Writer) (err error) {

	if err = writeLongstr(w, msg.Challenge); err != nil {
		return
	}

	return
}

// Read deserializes this message from the provided reader
func (msg *ConnectionSecure) Read(r io.Reader) (err error) {

	if msg.Challenge, err = readLongstr(r); err != nil {
		return
	}

	return
}

// ConnectionSecureOk represents the AMQP message connection.secure-ok
type ConnectionSecureOk struct {
	Response string
}

// ID returns the AMQP class and method identifiers for this message
func (msg *ConnectionSecureOk) ID() (uint16, uint16) {
	return 10, 21
}

// Wait returns true when the client should expect a response from the server
func (msg *ConnectionSecureOk) Wait() bool {
	return true
}

// Write serializes this message to the provided writer
func (msg *ConnectionSecureOk) Write(w io.Writer) (err error) {

	if err = writeLongstr(w, msg.Response); err != nil {
		return
	}

	return
}

// Read deserializes this message from the provided reader
func (msg *ConnectionSecureOk) Read(r io.Reader) (err error) {

	if msg.Response, err = readLongstr(r); err != nil {
		return
	}

	return
}

// ConnectionTune represents the AMQP message connection.tune
type ConnectionTune struct {
	ChannelMax uint16
	FrameMax   uint32
	Heartbeat  uint16
}

// ID returns the AMQP class and method identifiers for this message
func (msg *ConnectionTune) ID() (uint16, uint16) {
	return 10, 30
}

// Wait returns true when the client should expect a response from the server
func (msg *ConnectionTune) Wait() bool {
	return true
}

// Write serializes this message to the provided writer
func (msg *ConnectionTune) Write(w io.Writer) (err error) {

	if err = binary.Write(w, binary.BigEndian, msg.ChannelMax); err != nil {
		return
	}

	if err = binary.Write(w, binary.BigEndian, msg.FrameMax); err != nil {
		return
	}

	if err = binary.Write(w, binary.BigEndian, msg.Heartbeat); err != nil {
		return
	}

	return
}

// Read deserializes this message from the provided reader
func (msg *ConnectionTune) Read(r io.Reader) (err error) {

	if err = binary.Read(r, binary.BigEndian, &msg.ChannelMax); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &msg.FrameMax); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &msg.Heartbeat); err != nil {
		return
	}

	return
}

// ConnectionTuneOk represents the AMQP message connection.tune-ok
type ConnectionTuneOk struct {
	ChannelMax uint16
	FrameMax   uint32
	Heartbeat  uint16
}

// ID returns the AMQP class and method identifiers for this message
func (msg *ConnectionTuneOk) ID() (uint16, uint16) {
	return 10, 31
}

// Wait returns true when the client should expect a response from the server
func (msg *ConnectionTuneOk) Wait() bool {
	return true
}

// Write serializes this message to the provided writer
func (msg *ConnectionTuneOk) Write(w io.Writer) (err error) {

	if err = binary.Write(w, binary.BigEndian, msg.ChannelMax); err != nil {
		return
	}

	if err = binary.Write(w, binary.BigEndian, msg.FrameMax); err != nil {
		return
	}

	if err = binary.Write(w, binary.BigEndian, msg.Heartbeat); err != nil {
		return
	}

	return
}

// Read deserializes this message from the provided reader
func (msg *ConnectionTuneOk) Read(r io.Reader) (err error) {

	if err = binary.Read(r, binary.BigEndian, &msg.ChannelMax); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &msg.FrameMax); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &msg.Heartbeat); err != nil {
		return
	}

	return
}

// ConnectionOpen represents the AMQP message connection.open
type ConnectionOpen struct {
	VirtualHost string
	reserved1   string
	reserved2   bool
}

// ID returns the AMQP class and method identifiers for this message
func (msg *ConnectionOpen) ID() (uint16, uint16) {
	return 10, 40
}

// Wait returns true when the client should expect a response from the server
func (msg *ConnectionOpen) Wait() bool {
	return true
}

// Write serializes this message to the provided writer
func (msg *ConnectionOpen) Write(w io.Writer) (err error) {
	var bits byte

	if err = writeShortstr(w, msg.VirtualHost); err != nil {
		return
	}
	if err = writeShortstr(w, msg.reserved1); err != nil {
		return
	}

	if msg.reserved2 {
		bits |= 1 << 0
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	return
}

// Read deserializes this message from the provided reader
func (msg *ConnectionOpen) Read(r io.Reader) (err error) {
	var bits byte

	if msg.VirtualHost, err = readShortstr(r); err != nil {
		return
	}
	if msg.reserved1, err = readShortstr(r); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.reserved2 = (bits&(1<<0) > 0)

	return
}

// ConnectionOpenOk represents the AMQP message connection.open-ok
type ConnectionOpenOk struct {
	reserved1 string
}

// ID returns the AMQP class and method identifiers for this message
func (msg *ConnectionOpenOk) ID() (uint16, uint16) {
	return 10, 41
}

// Wait returns true when the client should expect a response from the server
func (msg *ConnectionOpenOk) Wait() bool {
	return true
}

// Write serializes this message to the provided writer
func (msg *ConnectionOpenOk) Write(w io.Writer) (err error) {

	if err = writeShortstr(w, msg.reserved1); err != nil {
		return
	}

	return
}

// Read deserializes this message from the provided reader
func (msg *ConnectionOpenOk) Read(r io.Reader) (err error) {

	if msg.reserved1, err = readShortstr(r); err != nil {
		return
	}

	return
}

// ConnectionClose represents the AMQP message connection.close
type ConnectionClose struct {
	ReplyCode uint16
	ReplyText string
	ClassId   uint16
	MethodId  uint16
}

// ID returns the AMQP class and method identifiers for this message
func (msg *ConnectionClose) ID() (uint16, uint16) {
	return 10, 50
}

// Wait returns true when the client should expect a response from the server
func (msg *ConnectionClose) Wait() bool {
	return true
}

// Write serializes this message to the provided writer
func (msg *ConnectionClose) Write(w io.Writer) (err error) {

	if err = binary.Write(w, binary.BigEndian, msg.ReplyCode); err != nil {
		return
	}

	if err = writeShortstr(w, msg.ReplyText); err != nil {
		return
	}

	if err = binary.Write(w, binary.BigEndian, msg.ClassId); err != nil {
		return
	}
	if err = binary.Write(w, binary.BigEndian, msg.MethodId); err != nil {
		return
	}

	return
}

// Read deserializes this message from the provided reader
func (msg *ConnectionClose) Read(r io.Reader) (err error) {

	if err = binary.Read(r, binary.BigEndian, &msg.ReplyCode); err != nil {
		return
	}

	if msg.ReplyText, err = readShortstr(r); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &msg.ClassId); err != nil {
		return
	}
	if err = binary.Read(r, binary.BigEndian, &msg.MethodId); err != nil {
		return
	}

	return
}

// ConnectionCloseOk represents the AMQP message connection.close-ok
type ConnectionCloseOk struct {
}

// ID returns the AMQP class and method identifiers for this message
func (msg *ConnectionCloseOk) ID() (uint16, uint16) {
	return 10, 51
}

// Wait returns true when the client should expect a response from the server
func (msg *ConnectionCloseOk) Wait() bool {
	return true
}

// Write serializes this message to the provided writer
func (msg *ConnectionCloseOk) Write(w io.Writer) (err error) {

	return
}

// Read deserializes this message from the provided reader
func (msg *ConnectionCloseOk) Read(r io.Reader) (err error) {

	return
}

// ConnectionBlocked represents the AMQP message connection.blocked
type ConnectionBlocked struct {
	Reason string
}

// ID returns the AMQP class and method identifiers for this message
func (msg *ConnectionBlocked) ID() (uint16, uint16) {
	return 10, 60
}

// Wait returns true when the client should expect a response from the server
func (msg *ConnectionBlocked) Wait() bool {
	return false
}

// Write serializes this message to the provided writer
func (msg *ConnectionBlocked) Write(w io.Writer) (err error) {

	if err = writeShortstr(w, msg.Reason); err != nil {
		return
	}

	return
}

// Read deserializes this message from the provided reader
func (msg *ConnectionBlocked) Read(r io.Reader) (err error) {

	if msg.Reason, err = readShortstr(r); err != nil {
		return
	}

	return
}

// ConnectionUnblocked represents the AMQP message connection.unblocked
type ConnectionUnblocked struct {
}

// ID returns the AMQP class and method identifiers for this message
func (msg *ConnectionUnblocked) ID() (uint16, uint16) {
	return 10, 61
}

// Wait returns true when the client should expect a response from the server
func (msg *ConnectionUnblocked) Wait() bool {
	return false
}

// Write serializes this message to the provided writer
func (msg *ConnectionUnblocked) Write(w io.Writer) (err error) {

	return
}

// Read deserializes this message from the provided reader
func (msg *ConnectionUnblocked) Read(r io.Reader) (err error) {

	return
}

// ChannelOpen represents the AMQP message channel.open
type ChannelOpen struct {
	reserved1 string
}

// ID returns the AMQP class and method identifiers for this message
func (msg *ChannelOpen) ID() (uint16, uint16) {
	return 20, 10
}

// Wait returns true when the client should expect a response from the server
func (msg *ChannelOpen) Wait() bool {
	return true
}

// Write serializes this message to the provided writer
func (msg *ChannelOpen) Write(w io.Writer) (err error) {

	if err = writeShortstr(w, msg.reserved1); err != nil {
		return
	}

	return
}

// Read deserializes this message from the provided reader
func (msg *ChannelOpen) Read(r io.Reader) (err error) {

	if msg.reserved1, err = readShortstr(r); err != nil {
		return
	}

	return
}

// ChannelOpenOk represents the AMQP message channel.open-ok
type ChannelOpenOk struct {
	reserved1 string
}

// ID returns the AMQP class and method identifiers for this message
func (msg *ChannelOpenOk) ID() (uint16, uint16) {
	return 20, 11
}

// Wait returns true when the client should expect a response from the server
func (msg *ChannelOpenOk) Wait() bool {
	return true
}

// Write serializes this message to the provided writer
func (msg *ChannelOpenOk) Write(w io.Writer) (err error) {

	if err = writeLongstr(w, msg.reserved1); err != nil {
		return
	}

	return
}

// Read deserializes this message from the provided reader
func (msg *ChannelOpenOk) Read(r io.Reader) (err error) {

	if msg.reserved1, err = readLongstr(r); err != nil {
		return
	}

	return
}

// ChannelFlow represents the AMQP message channel.flow
type ChannelFlow struct {
	Active bool
}

// ID returns the AMQP class and method identifiers for this message
func (msg *ChannelFlow) ID() (uint16, uint16) {
	return 20, 20
}

// Wait returns true when the client should expect a response from the server
func (msg *ChannelFlow) Wait() bool {
	return true
}

// Write serializes this message to the provided writer
func (msg *ChannelFlow) Write(w io.Writer) (err error) {
	var bits byte

	if msg.Active {
		bits |= 1 << 0
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	return
}

// Read deserializes this message from the provided reader
func (msg *ChannelFlow) Read(r io.Reader) (err error) {
	var bits byte

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.Active = (bits&(1<<0) > 0)

	return
}

// ChannelFlowOk represents the AMQP message channel.flow-ok
type ChannelFlowOk struct {
	Active bool
}

// ID returns the AMQP class and method identifiers for this message
func (msg *ChannelFlowOk) ID() (uint16, uint16) {
	return 20, 21
}

// Wait returns true when the client should expect a response from the server
func (msg *ChannelFlowOk) Wait() bool {
	return false
}

// Write serializes this message to the provided writer
func (msg *ChannelFlowOk) Write(w io.Writer) (err error) {
	var bits byte

	if msg.Active {
		bits |= 1 << 0
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	return
}

// Read deserializes this message from the provided reader
func (msg *ChannelFlowOk) Read(r io.Reader) (err error) {
	var bits byte

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.Active = (bits&(1<<0) > 0)

	return
}

// ChannelClose represents the AMQP message channel.close
type ChannelClose struct {
	ReplyCode uint16
	ReplyText string
	ClassId   uint16
	MethodId  uint16
}

// ID returns the AMQP class and method identifiers for this message
func (msg *ChannelClose) ID() (uint16, uint16) {
	return 20, 40
}

// Wait returns true when the client should expect a response from the server
func (msg *ChannelClose) Wait() bool {
	return true
}

// Write serializes this message to the provided writer
func (msg *ChannelClose) Write(w io.Writer) (err error) {

	if err = binary.Write(w, binary.BigEndian, msg.ReplyCode); err != nil {
		return
	}

	if err = writeShortstr(w, msg.ReplyText); err != nil {
		return
	}

	if err = binary.Write(w, binary.BigEndian, msg.ClassId); err != nil {
		return
	}
	if err = binary.Write(w, binary.BigEndian, msg.MethodId); err != nil {
		return
	}

	return
}

// Read deserializes this message from the provided reader
func (msg *ChannelClose) Read(r io.Reader) (err error) {

	if err = binary.Read(r, binary.BigEndian, &msg.ReplyCode); err != nil {
		return
	}

	if msg.ReplyText, err = readShortstr(r); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &msg.ClassId); err != nil {
		return
	}
	if err = binary.Read(r, binary.BigEndian, &msg.MethodId); err != nil {
		return
	}

	return
}

// ChannelCloseOk represents the AMQP message channel.close-ok
type ChannelCloseOk struct {
}

// ID returns the AMQP class and method identifiers for this message
func (msg *ChannelCloseOk) ID() (uint16, uint16) {
	return 20, 41
}

// Wait returns true when the client should expect a response from the server
func (msg *ChannelCloseOk) Wait() bool {
	return true
}

// Write serializes this message to the provided writer
func (msg *ChannelCloseOk) Write(w io.Writer) (err error) {

	return
}

// Read deserializes this message from the provided reader
func (msg *ChannelCloseOk) Read(r io.Reader) (err error) {

	return
}

// ExchangeDeclare represents the AMQP message exchange.declare
type ExchangeDeclare struct {
	reserved1  uint16
	Exchange   string
	Type       string
	Passive    bool
	Durable    bool
	AutoDelete bool
	Internal   bool
	NoWait     bool
	Arguments  Table
}

// ID returns the AMQP class and method identifiers for this message
func (msg *ExchangeDeclare) ID() (uint16, uint16) {
	return 40, 10
}

// Wait returns true when the client should expect a response from the server
func (msg *ExchangeDeclare) Wait() bool {
	return true && !msg.NoWait
}

// Write serializes this message to the provided writer
func (msg *ExchangeDeclare) Write(w io.Writer) (err error) {
	var bits byte

	if err = binary.Write(w, binary.BigEndian, msg.reserved1); err != nil {
		return
	}

	if err = writeShortstr(w, msg.Exchange); err != nil {
		return
	}
	if err = writeShortstr(w, msg.Type); err != nil {
		return
	}

	if msg.Passive {
		bits |= 1 << 0
	}

	if msg.Durable {
		bits |= 1 << 1
	}

	if msg.AutoDelete {
		bits |= 1 << 2
	}

	if msg.Internal {
		bits |= 1 << 3
	}

	if msg.NoWait {
		bits |= 1 << 4
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	if err = writeTable(w, msg.Arguments); err != nil {
		return
	}

	return
}

// Read deserializes this message from the provided reader
func (msg *ExchangeDeclare) Read(r io.Reader) (err error) {
	var bits byte

	if err = binary.Read(r, binary.BigEndian, &msg.reserved1); err != nil {
		return
	}

	if msg.Exchange, err = readShortstr(r); err != nil {
		return
	}
	if msg.Type, err = readShortstr(r); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.Passive = (bits&(1<<0) > 0)
	msg.Durable = (bits&(1<<1) > 0)
	msg.AutoDelete = (bits&(1<<2) > 0)
	msg.Internal = (bits&(1<<3) > 0)
	msg.NoWait = (bits&(1<<4) > 0)

	if msg.Arguments, err = readTable(r); err != nil {
		return
	}

	return
}

// ExchangeDeclareOk represents the AMQP message exchange.declare-ok
type ExchangeDeclareOk struct {
}

// ID returns the AMQP class and method identifiers for this message
func (msg *ExchangeDeclareOk) ID() (uint16, uint16) {
	return 40, 11
}

// Wait returns true when the client should expect a response from the server
func (msg *ExchangeDeclareOk) Wait() bool {
	return true
}

// Write serializes this message to the provided writer
func (msg *ExchangeDeclareOk) Write(w io.Writer) (err error) {

	return
}

// Read deserializes this message from the provided reader
func (msg *ExchangeDeclareOk) Read(r io.Reader) (err error) {

	return
}

// ExchangeDelete represents the AMQP message exchange.delete
type ExchangeDelete struct {
	reserved1 uint16
	Exchange  string
	IfUnused  bool
	NoWait    bool
}

// ID returns the AMQP class and method identifiers for this message
func (msg *ExchangeDelete) ID() (uint16, uint16) {
	return 40, 20
}

// Wait returns true when the client should expect a response from the server
func (msg *ExchangeDelete) Wait() bool {
	return true && !msg.NoWait
}

// Write serializes this message to the provided writer
func (msg *ExchangeDelete) Write(w io.Writer) (err error) {
	var bits byte

	if err = binary.Write(w, binary.BigEndian, msg.reserved1); err != nil {
		return
	}

	if err = writeShortstr(w, msg.Exchange); err != nil {
		return
	}

	if msg.IfUnused {
		bits |= 1 << 0
	}

	if msg.NoWait {
		bits |= 1 << 1
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	return
}

// Read deserializes this message from the provided reader
func (msg *ExchangeDelete) Read(r io.Reader) (err error) {
	var bits byte

	if err = binary.Read(r, binary.BigEndian, &msg.reserved1); err != nil {
		return
	}

	if msg.Exchange, err = readShortstr(r); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.IfUnused = (bits&(1<<0) > 0)
	msg.NoWait = (bits&(1<<1) > 0)

	return
}

// ExchangeDeleteOk represents the AMQP message exchange.delete-ok
type ExchangeDeleteOk struct {
}

// ID returns the AMQP class and method identifiers for this message
func (msg *ExchangeDeleteOk) ID() (uint16, uint16) {
	return 40, 21
}

// Wait returns true when the client should expect a response from the server
func (msg *ExchangeDeleteOk) Wait() bool {
	return true
}

// Write serializes this message to the provided writer
func (msg *ExchangeDeleteOk) Write(w io.Writer) (err error) {

	return
}

// Read deserializes this message from the provided reader
func (msg *ExchangeDeleteOk) Read(r io.Reader) (err error) {

	return
}

// ExchangeBind represents the AMQP message exchange.bind
type ExchangeBind struct {
	reserved1   uint16
	Destination string
	Source      string
	RoutingKey  string
	NoWait      bool
	Arguments   Table
}

// ID returns the AMQP class and method identifiers for this message
func (msg *ExchangeBind) ID() (uint16, uint16) {
	return 40, 30
}

// Wait returns true when the client should expect a response from the server
func (msg *ExchangeBind) Wait() bool {
	return true && !msg.NoWait
}

// Write serializes this message to the provided writer
func (msg *ExchangeBind) Write(w io.Writer) (err error) {
	var bits byte

	if err = binary.Write(w, binary.BigEndian, msg.reserved1); err != nil {
		return
	}

	if err = writeShortstr(w, msg.Destination); err != nil {
		return
	}
	if err = writeShortstr(w, msg.Source); err != nil {
		return
	}
	if err = writeShortstr(w, msg.RoutingKey); err != nil {
		return
	}

	if msg.NoWait {
		bits |= 1 << 0
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	if err = writeTable(w, msg.Arguments); err != nil {
		return
	}

	return
}

// Read deserializes this message from the provided reader
func (msg *ExchangeBind) Read(r io.Reader) (err error) {
	var bits byte

	if err = binary.Read(r, binary.BigEndian, &msg.reserved1); err != nil {
		return
	}

	if msg.Destination, err = readShortstr(r); err != nil {
		return
	}
	if msg.Source, err = readShortstr(r); err != nil {
		return
	}
	if msg.RoutingKey, err = readShortstr(r); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.NoWait = (bits&(1<<0) > 0)

	if msg.Arguments, err = readTable(r); err != nil {
		return
	}

	return
}

// ExchangeBindOk represents the AMQP message exchange.bind-ok
type ExchangeBindOk struct {
}

// ID returns the AMQP class and method identifiers for this message
func (msg *ExchangeBindOk) ID() (uint16, uint16) {
	return 40, 31
}

// Wait returns true when the client should expect a response from the server
func (msg *ExchangeBindOk) Wait() bool {
	return true
}

// Write serializes this message to the provided writer
func (msg *ExchangeBindOk) Write(w io.Writer) (err error) {

	return
}

// Read deserializes this message from the provided reader
func (msg *ExchangeBindOk) Read(r io.Reader) (err error) {

	return
}

// ExchangeUnbind represents the AMQP message exchange.unbind
type ExchangeUnbind struct {
	reserved1   uint16
	Destination string
	Source      string
	RoutingKey  string
	NoWait      bool
	Arguments   Table
}

// ID returns the AMQP class and method identifiers for this message
func (msg *ExchangeUnbind) ID() (uint16, uint16) {
	return 40, 40
}

// Wait returns true when the client should expect a response from the server
func (msg *ExchangeUnbind) Wait() bool {
	return true && !msg.NoWait
}

// Write serializes this message to the provided writer
func (msg *ExchangeUnbind) Write(w io.Writer) (err error) {
	var bits byte

	if err = binary.Write(w, binary.BigEndian, msg.reserved1); err != nil {
		return
	}

	if err = writeShortstr(w, msg.Destination); err != nil {
		return
	}
	if err = writeShortstr(w, msg.Source); err != nil {
		return
	}
	if err = writeShortstr(w, msg.RoutingKey); err != nil {
		return
	}

	if msg.NoWait {
		bits |= 1 << 0
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	if err = writeTable(w, msg.Arguments); err != nil {
		return
	}

	return
}

// Read deserializes this message from the provided reader
func (msg *ExchangeUnbind) Read(r io.Reader) (err error) {
	var bits byte

	if err = binary.Read(r, binary.BigEndian, &msg.reserved1); err != nil {
		return
	}

	if msg.Destination, err = readShortstr(r); err != nil {
		return
	}
	if msg.Source, err = readShortstr(r); err != nil {
		return
	}
	if msg.RoutingKey, err = readShortstr(r); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.NoWait = (bits&(1<<0) > 0)

	if msg.Arguments, err = readTable(r); err != nil {
		return
	}

	return
}

// ExchangeUnbindOk represents the AMQP message exchange.unbind-ok
type ExchangeUnbindOk struct {
}

// ID returns the AMQP class and method identifiers for this message
func (msg *ExchangeUnbindOk) ID() (uint16, uint16) {
	return 40, 51
}

// Wait returns true when the client should expect a response from the server
func (msg *ExchangeUnbindOk) Wait() bool {
	return true
}

// Write serializes this message to the provided writer
func (msg *ExchangeUnbindOk) Write(w io.Writer) (err error) {

	return
}

// Read deserializes this message from the provided reader
func (msg *ExchangeUnbindOk) Read(r io.Reader) (err error) {

	return
}

// QueueDeclare represents the AMQP message queue.declare
type QueueDeclare struct {
	reserved1  uint16
	Queue      string
	Passive    bool
	Durable    bool
	Exclusive  bool
	AutoDelete bool
	NoWait     bool
	Arguments  Table
}

// ID returns the AMQP class and method identifiers for this message
func (msg *QueueDeclare) ID() (uint16, uint16) {
	return 50, 10
}

// Wait returns true when the client should expect a response from the server
func (msg *QueueDeclare) Wait() bool {
	return true && !msg.NoWait
}

// Write serializes this message to the provided writer
func (msg *QueueDeclare) Write(w io.Writer) (err error) {
	var bits byte

	if err = binary.Write(w, binary.BigEndian, msg.reserved1); err != nil {
		return
	}

	if err = writeShortstr(w, msg.Queue); err != nil {
		return
	}

	if msg.Passive {
		bits |= 1 << 0
	}

	if msg.Durable {
		bits |= 1 << 1
	}

	if msg.Exclusive {
		bits |= 1 << 2
	}

	if msg.AutoDelete {
		bits |= 1 << 3
	}

	if msg.NoWait {
		bits |= 1 << 4
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	if err = writeTable(w, msg.Arguments); err != nil {
		return
	}

	return
}

// Read deserializes this message from the provided reader
func (msg *QueueDeclare) Read(r io.Reader) (err error) {
	var bits byte

	if err = binary.Read(r, binary.BigEndian, &msg.reserved1); err != nil {
		return
	}

	if msg.Queue, err = readShortstr(r); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.Passive = (bits&(1<<0) > 0)
	msg.Durable = (bits&(1<<1) > 0)
	msg.Exclusive = (bits&(1<<2) > 0)
	msg.AutoDelete = (bits&(1<<3) > 0)
	msg.NoWait = (bits&(1<<4) > 0)

	if msg.Arguments, err = readTable(r); err != nil {
		return
	}

	return
}

// QueueDeclareOk represents the AMQP message queue.declare-ok
type QueueDeclareOk struct {
	Queue         string
	MessageCount  uint32
	ConsumerCount uint32
}

// ID returns the AMQP class and method identifiers for this message
func (msg *QueueDeclareOk) ID() (uint16, uint16) {
	return 50, 11
}

// Wait returns true when the client should expect a response from the server
func (msg *QueueDeclareOk) Wait() bool {
	return true
}

// Write serializes this message to the provided writer
func (msg *QueueDeclareOk) Write(w io.Writer) (err error) {

	if err = writeShortstr(w, msg.Queue); err != nil {
		return
	}

	if err = binary.Write(w, binary.BigEndian, msg.MessageCount); err != nil {
		return
	}
	if err = binary.Write(w, binary.BigEndian, msg.ConsumerCount); err != nil {
		return
	}

	return
}

// Read deserializes this message from the provided reader
func (msg *QueueDeclareOk) Read(r io.Reader) (err error) {

	if msg.Queue, err = readShortstr(r); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &msg.MessageCount); err != nil {
		return
	}
	if err = binary.Read(r, binary.BigEndian, &msg.ConsumerCount); err != nil {
		return
	}

	return
}

// QueueBind represents the AMQP message queue.bind
type QueueBind struct {
	reserved1  uint16
	Queue      string
	Exchange   string
	RoutingKey string
	NoWait     bool
	Arguments  Table
}

// ID returns the AMQP class and method identifiers for this message
func (msg *QueueBind) ID() (uint16, uint16) {
	return 50, 20
}

// Wait returns true when the client should expect a response from the server
func (msg *QueueBind) Wait() bool {
	return true && !msg.NoWait
}

// Write serializes this message to the provided writer
func (msg *QueueBind) Write(w io.Writer) (err error) {
	var bits byte

	if err = binary.Write(w, binary.BigEndian, msg.reserved1); err != nil {
		return
	}

	if err = writeShortstr(w, msg.Queue); err != nil {
		return
	}
	if err = writeShortstr(w, msg.Exchange); err != nil {
		return
	}
	if err = writeShortstr(w, msg.RoutingKey); err != nil {
		return
	}

	if msg.NoWait {
		bits |= 1 << 0
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	if err = writeTable(w, msg.Arguments); err != nil {
		return
	}

	return
}

// Read deserializes this message from the provided reader
func (msg *QueueBind) Read(r io.Reader) (err error) {
	var bits byte

	if err = binary.Read(r, binary.BigEndian, &msg.reserved1); err != nil {
		return
	}

	if msg.Queue, err = readShortstr(r); err != nil {
		return
	}
	if msg.Exchange, err = readShortstr(r); err != nil {
		return
	}
	if msg.RoutingKey, err = readShortstr(r); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.NoWait = (bits&(1<<0) > 0)

	if msg.Arguments, err = readTable(r); err != nil {
		return
	}

	return
}

// QueueBindOk represents the AMQP message queue.bind-ok
type QueueBindOk struct {
}

// ID returns the AMQP class and method identifiers for this message
func (msg *QueueBindOk) ID() (uint16, uint16) {
	return 50, 21
}

// Wait returns true when the client should expect a response from the server
func (msg *QueueBindOk) Wait() bool {
	return true
}

// Write serializes this message to the provided writer
func (msg *QueueBindOk) Write(w io.Writer) (err error) {

	return
}

// Read deserializes this message from the provided reader
func (msg *QueueBindOk) Read(r io.Reader) (err error) {

	return
}

// QueueUnbind represents the AMQP message queue.unbind
type QueueUnbind struct {
	reserved1  uint16
	Queue      string
	Exchange   string
	RoutingKey string
	Arguments  Table
}

// ID returns the AMQP class and method identifiers for this message
func (msg *QueueUnbind) ID() (uint16, uint16) {
	return 50, 50
}

// Wait returns true when the client should expect a response from the server
func (msg *QueueUnbind) Wait() bool {
	return true
}

// Write serializes this message to the provided writer
func (msg *QueueUnbind) Write(w io.Writer) (err error) {

	if err = binary.Write(w, binary.BigEndian, msg.reserved1); err != nil {
		return
	}

	if err = writeShortstr(w, msg.Queue); err != nil {
		return
	}
	if err = writeShortstr(w, msg.Exchange); err != nil {
		return
	}
	if err = writeShortstr(w, msg.RoutingKey); err != nil {
		return
	}

	if err = writeTable(w, msg.Arguments); err != nil {
		return
	}

	return
}

// Read deserializes this message from the provided reader
func (msg *QueueUnbind) Read(r io.Reader) (err error) {

	if err = binary.Read(r, binary.BigEndian, &msg.reserved1); err != nil {
		return
	}

	if msg.Queue, err = readShortstr(r); err != nil {
		return
	}
	if msg.Exchange, err = readShortstr(r); err != nil {
		return
	}
	if msg.RoutingKey, err = readShortstr(r); err != nil {
		return
	}

	if msg.Arguments, err = readTable(r); err != nil {
		return
	}

	return
}

// QueueUnbindOk represents the AMQP message queue.unbind-ok
type QueueUnbindOk struct {
}

// ID returns the AMQP class and method identifiers for this message
func (msg *QueueUnbindOk) ID() (uint16, uint16) {
	return 50, 51
}

// Wait returns true when the client should expect a response from the server
func (msg *QueueUnbindOk) Wait() bool {
	return true
}

// Write serializes this message to the provided writer
func (msg *QueueUnbindOk) Write(w io.Writer) (err error) {

	return
}

// Read deserializes this message from the provided reader
func (msg *QueueUnbindOk) Read(r io.Reader) (err error) {

	return
}

// QueuePurge represents the AMQP message queue.purge
type QueuePurge struct {
	reserved1 uint16
	Queue     string
	NoWait    bool
}

// ID returns the AMQP class and method identifiers for this message
func (msg *QueuePurge) ID() (uint16, uint16) {
	return 50, 30
}

// Wait returns true when the client should expect a response from the server
func (msg *QueuePurge) Wait() bool {
	return true && !msg.NoWait
}

// Write serializes this message to the provided writer
func (msg *QueuePurge) Write(w io.Writer) (err error) {
	var bits byte

	if err = binary.Write(w, binary.BigEndian, msg.reserved1); err != nil {
		return
	}

	if err = writeShortstr(w, msg.Queue); err != nil {
		return
	}

	if msg.NoWait {
		bits |= 1 << 0
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	return
}

// Read deserializes this message from the provided reader
func (msg *QueuePurge) Read(r io.Reader) (err error) {
	var bits byte

	if err = binary.Read(r, binary.BigEndian, &msg.reserved1); err != nil {
		return
	}

	if msg.Queue, err = readShortstr(r); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.NoWait = (bits&(1<<0) > 0)

	return
}

// QueuePurgeOk represents the AMQP message queue.purge-ok
type QueuePurgeOk struct {
	MessageCount uint32
}

// ID returns the AMQP class and method identifiers for this message
func (msg *QueuePurgeOk) ID() (uint16, uint16) {
	return 50, 31
}

// Wait returns true when the client should expect a response from the server
func (msg *QueuePurgeOk) Wait() bool {
	return true
}

// Write serializes this message to the provided writer
func (msg *QueuePurgeOk) Write(w io.Writer) (err error) {

	if err = binary.Write(w, binary.BigEndian, msg.MessageCount); err != nil {
		return
	}

	return
}

// Read deserializes this message from the provided reader
func (msg *QueuePurgeOk) Read(r io.Reader) (err error) {

	if err = binary.Read(r, binary.BigEndian, &msg.MessageCount); err != nil {
		return
	}

	return
}

// QueueDelete represents the AMQP message queue.delete
type QueueDelete struct {
	reserved1 uint16
	Queue     string
	IfUnused  bool
	IfEmpty   bool
	NoWait    bool
}

// ID returns the AMQP class and method identifiers for this message
func (msg *QueueDelete) ID() (uint16, uint16) {
	return 50, 40
}

// Wait returns true when the client should expect a response from the server
func (msg *QueueDelete) Wait() bool {
	return true && !msg.NoWait
}

// Write serializes this message to the provided writer
func (msg *QueueDelete) Write(w io.Writer) (err error) {
	var bits byte

	if err = binary.Write(w, binary.BigEndian, msg.reserved1); err != nil {
		return
	}

	if err = writeShortstr(w, msg.Queue); err != nil {
		return
	}

	if msg.IfUnused {
		bits |= 1 << 0
	}

	if msg.IfEmpty {
		bits |= 1 << 1
	}

	if msg.NoWait {
		bits |= 1 << 2
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	return
}

// Read deserializes this message from the provided reader
func (msg *QueueDelete) Read(r io.Reader) (err error) {
	var bits byte

	if err = binary.Read(r, binary.BigEndian, &msg.reserved1); err != nil {
		return
	}

	if msg.Queue, err = readShortstr(r); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.IfUnused = (bits&(1<<0) > 0)
	msg.IfEmpty = (bits&(1<<1) > 0)
	msg.NoWait = (bits&(1<<2) > 0)

	return
}

// QueueDeleteOk represents the AMQP message queue.delete-ok
type QueueDeleteOk struct {
	MessageCount uint32
}

// ID returns the AMQP class and method identifiers for this message
func (msg *QueueDeleteOk) ID() (uint16, uint16) {
	return 50, 41
}

// Wait returns true when the client should expect a response from the server
func (msg *QueueDeleteOk) Wait() bool {
	return true
}

// Write serializes this message to the provided writer
func (msg *QueueDeleteOk) Write(w io.Writer) (err error) {

	if err = binary.Write(w, binary.BigEndian, msg.MessageCount); err != nil {
		return
	}

	return
}

// Read deserializes this message from the provided reader
func (msg *QueueDeleteOk) Read(r io.Reader) (err error) {

	if err = binary.Read(r, binary.BigEndian, &msg.MessageCount); err != nil {
		return
	}

	return
}

// BasicQos represents the AMQP message basic.qos
type BasicQos struct {
	PrefetchSize  uint32
	PrefetchCount uint16
	Global        bool
}

// ID returns the AMQP class and method identifiers for this message
func (msg *BasicQos) ID() (uint16, uint16) {
	return 60, 10
}

// Wait returns true when the client should expect a response from the server
func (msg *BasicQos) Wait() bool {
	return true
}

// Write serializes this message to the provided writer
func (msg *BasicQos) Write(w io.Writer) (err error) {
	var bits byte

	if err = binary.Write(w, binary.BigEndian, msg.PrefetchSize); err != nil {
		return
	}

	if err = binary.Write(w, binary.BigEndian, msg.PrefetchCount); err != nil {
		return
	}

	if msg.Global {
		bits |= 1 << 0
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	return
}

// Read deserializes this message from the provided reader
func (msg *BasicQos) Read(r io.Reader) (err error) {
	var bits byte

	if err = binary.Read(r, binary.BigEndian, &msg.PrefetchSize); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &msg.PrefetchCount); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.Global = (bits&(1<<0) > 0)

	return
}

// BasicQosOk represents the AMQP message basic.qos-ok
type BasicQosOk struct {
}

// ID returns the AMQP class and method identifiers for this message
func (msg *BasicQosOk) ID() (uint16, uint16) {
	return 60, 11
}

// Wait returns true when the client should expect a response from the server
func (msg *BasicQosOk) Wait() bool {
	return true
}

// Write serializes this message to the provided writer
func (msg *BasicQosOk) Write(w io.Writer) (err error) {

	return
}

// Read deserializes this message from the provided reader
func (msg *BasicQosOk) Read(r io.Reader) (err error) {

	return
}

// BasicConsume represents the AMQP message basic.consume
type BasicConsume struct {
	reserved1   uint16
	Queue       string
	ConsumerTag string
	NoLocal     bool
	NoAck       bool
	Exclusive   bool
	NoWait      bool
	Arguments   Table
}

// ID returns the AMQP class and method identifiers for this message
func (msg *BasicConsume) ID() (uint16, uint16) {
	return 60, 20
}

// Wait returns true when the client should expect a response from the server
func (msg *BasicConsume) Wait() bool {
	return true && !msg.NoWait
}

// Write serializes this message to the provided writer
func (msg *BasicConsume) Write(w io.Writer) (err error) {
	var bits byte

	if err = binary.Write(w, binary.BigEndian, msg.reserved1); err != nil {
		return
	}

	if err = writeShortstr(w, msg.Queue); err != nil {
		return
	}
	if err = writeShortstr(w, msg.ConsumerTag); err != nil {
		return
	}

	if msg.NoLocal {
		bits |= 1 << 0
	}

	if msg.NoAck {
		bits |= 1 << 1
	}

	if msg.Exclusive {
		bits |= 1 << 2
	}

	if msg.NoWait {
		bits |= 1 << 3
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	if err = writeTable(w, msg.Arguments); err != nil {
		return
	}

	return
}

// Read deserializes this message from the provided reader
func (msg *BasicConsume) Read(r io.Reader) (err error) {
	var bits byte

	if err = binary.Read(r, binary.BigEndian, &msg.reserved1); err != nil {
		return
	}

	if msg.Queue, err = readShortstr(r); err != nil {
		return
	}
	if msg.ConsumerTag, err = readShortstr(r); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.NoLocal = (bits&(1<<0) > 0)
	msg.NoAck = (bits&(1<<1) > 0)
	msg.Exclusive = (bits&(1<<2) > 0)
	msg.NoWait = (bits&(1<<3) > 0)

	if msg.Arguments, err = readTable(r); err != nil {
		return
	}

	return
}

// BasicConsumeOk represents the AMQP message basic.consume-ok
type BasicConsumeOk struct {
	ConsumerTag string
}

// ID returns the AMQP class and method identifiers for this message
func (msg *BasicConsumeOk) ID() (uint16, uint16) {
	return 60, 21
}

// Wait returns true when the client should expect a response from the server
func (msg *BasicConsumeOk) Wait() bool {
	return true
}

// Write serializes this message to the provided writer
func (msg *BasicConsumeOk) Write(w io.Writer) (err error) {

	if err = writeShortstr(w, msg.ConsumerTag); err != nil {
		return
	}

	return
}

// Read deserializes this message from the provided reader
func (msg *BasicConsumeOk) Read(r io.Reader) (err error) {

	if msg.ConsumerTag, err = readShortstr(r); err != nil {
		return
	}

	return
}

// BasicCancel represents the AMQP message basic.cancel
type BasicCancel struct {
	ConsumerTag string
	NoWait      bool
}

// ID returns the AMQP class and method identifiers for this message
func (msg *BasicCancel) ID() (uint16, uint16) {
	return 60, 30
}

// Wait returns true when the client should expect a response from the server
func (msg *BasicCancel) Wait() bool {
	return true && !msg.NoWait
}

// Write serializes this message to the provided writer
func (msg *BasicCancel) Write(w io.Writer) (err error) {
	var bits byte

	if err = writeShortstr(w, msg.ConsumerTag); err != nil {
		return
	}

	if msg.NoWait {
		bits |= 1 << 0
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	return
}

// Read deserializes this message from the provided reader
func (msg *BasicCancel) Read(r io.Reader) (err error) {
	var bits byte

	if msg.ConsumerTag, err = readShortstr(r); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.NoWait = (bits&(1<<0) > 0)

	return
}

// BasicCancelOk represents the AMQP message basic.cancel-ok
type BasicCancelOk struct {
	ConsumerTag string
}

// ID returns the AMQP class and method identifiers for this message
func (msg *BasicCancelOk) ID() (uint16, uint16) {
	return 60, 31
}

// Wait returns true when the client should expect a response from the server
func (msg *BasicCancelOk) Wait() bool {
	return true
}

// Write serializes this message to the provided writer
func (msg *BasicCancelOk) Write(w io.Writer) (err error) {

	if err = writeShortstr(w, msg.ConsumerTag); err != nil {
		return
	}

	return
}

// Read deserializes this message from the provided reader
func (msg *BasicCancelOk) Read(r io.Reader) (err error) {

	if msg.ConsumerTag, err = readShortstr(r); err != nil {
		return
	}

	return
}

// BasicPublish represents the AMQP message basic.publish
type BasicPublish struct {
	reserved1  uint16
	Exchange   string
	RoutingKey string
	Mandatory  bool
	Immediate  bool
	Properties Properties
	Body       []byte
}

// ID returns the AMQP class and method identifiers for this message
func (msg *BasicPublish) ID() (uint16, uint16) {
	return 60, 40
}

// Wait returns true when the client should expect a response from the server
func (msg *BasicPublish) Wait() bool {
	return false
}

// GetContent returns the Properties and Body from this message
func (msg *BasicPublish) GetContent() (Properties, []byte) {
	return msg.Properties, msg.Body
}

// SetContent sets the Properties and Body for serialization
func (msg *BasicPublish) SetContent(props Properties, body []byte) {
	msg.Properties, msg.Body = props, body
}

// Write serializes this message to the provided writer
func (msg *BasicPublish) Write(w io.Writer) (err error) {
	var bits byte

	if err = binary.Write(w, binary.BigEndian, msg.reserved1); err != nil {
		return
	}

	if err = writeShortstr(w, msg.Exchange); err != nil {
		return
	}
	if err = writeShortstr(w, msg.RoutingKey); err != nil {
		return
	}

	if msg.Mandatory {
		bits |= 1 << 0
	}

	if msg.Immediate {
		bits |= 1 << 1
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	return
}

// Read deserializes this message from the provided reader
func (msg *BasicPublish) Read(r io.Reader) (err error) {
	var bits byte

	if err = binary.Read(r, binary.BigEndian, &msg.reserved1); err != nil {
		return
	}

	if msg.Exchange, err = readShortstr(r); err != nil {
		return
	}
	if msg.RoutingKey, err = readShortstr(r); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.Mandatory = (bits&(1<<0) > 0)
	msg.Immediate = (bits&(1<<1) > 0)

	return
}

// BasicReturn represents the AMQP message basic.return
type BasicReturn struct {
	ReplyCode  uint16
	ReplyText  string
	Exchange   string
	RoutingKey string
	Properties Properties
	Body       []byte
}

// ID returns the AMQP class and method identifiers for this message
func (msg *BasicReturn) ID() (uint16, uint16) {
	return 60, 50
}

// Wait returns true when the client should expect a response from the server
func (msg *BasicReturn) Wait() bool {
	return false
}

// GetContent returns the Properties and Body from this message
func (msg *BasicReturn) GetContent() (Properties, []byte) {
	return msg.Properties, msg.Body
}

// SetContent sets the Properties and Body for serialization
func (msg *BasicReturn) SetContent(props Properties, body []byte) {
	msg.Properties, msg.Body = props, body
}

// Write serializes this message to the provided writer
func (msg *BasicReturn) Write(w io.Writer) (err error) {

	if err = binary.Write(w, binary.BigEndian, msg.ReplyCode); err != nil {
		return
	}

	if err = writeShortstr(w, msg.ReplyText); err != nil {
		return
	}
	if err = writeShortstr(w, msg.Exchange); err != nil {
		return
	}
	if err = writeShortstr(w, msg.RoutingKey); err != nil {
		return
	}

	return
}

// Read deserializes this message from the provided reader
func (msg *BasicReturn) Read(r io.Reader) (err error) {

	if err = binary.Read(r, binary.BigEndian, &msg.ReplyCode); err != nil {
		return
	}

	if msg.ReplyText, err = readShortstr(r); err != nil {
		return
	}
	if msg.Exchange, err = readShortstr(r); err != nil {
		return
	}
	if msg.RoutingKey, err = readShortstr(r); err != nil {
		return
	}

	return
}

// BasicDeliver represents the AMQP message basic.deliver
type BasicDeliver struct {
	ConsumerTag string
	DeliveryTag uint64
	Redelivered bool
	Exchange    string
	RoutingKey  string
	Properties  Properties
	Body        []byte
}

// ID returns the AMQP class and method identifiers for this message
func (msg *BasicDeliver) ID() (uint16, uint16) {
	return 60, 60
}

// Wait returns true when the client should expect a response from the server
func (msg *BasicDeliver) Wait() bool {
	return false
}

// GetContent returns the Properties and Body from this message
func (msg *BasicDeliver) GetContent() (Properties, []byte) {
	return msg.Properties, msg.Body
}

// SetContent sets the Properties and Body for serialization
func (msg *BasicDeliver) SetContent(props Properties, body []byte) {
	msg.Properties, msg.Body = props, body
}

// Write serializes this message to the provided writer
func (msg *BasicDeliver) Write(w io.Writer) (err error) {
	var bits byte

	if err = writeShortstr(w, msg.ConsumerTag); err != nil {
		return
	}

	if err = binary.Write(w, binary.BigEndian, msg.DeliveryTag); err != nil {
		return
	}

	if msg.Redelivered {
		bits |= 1 << 0
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	if err = writeShortstr(w, msg.Exchange); err != nil {
		return
	}
	if err = writeShortstr(w, msg.RoutingKey); err != nil {
		return
	}

	return
}

// Read deserializes this message from the provided reader
func (msg *BasicDeliver) Read(r io.Reader) (err error) {
	var bits byte

	if msg.ConsumerTag, err = readShortstr(r); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &msg.DeliveryTag); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.Redelivered = (bits&(1<<0) > 0)

	if msg.Exchange, err = readShortstr(r); err != nil {
		return
	}
	if msg.RoutingKey, err = readShortstr(r); err != nil {
		return
	}

	return
}

// BasicGet represents the AMQP message basic.get
type BasicGet struct {
	reserved1 uint16
	Queue     string
	NoAck     bool
}

// ID returns the AMQP class and method identifiers for this message
func (msg *BasicGet) ID() (uint16, uint16) {
	return 60, 70
}

// Wait returns true when the client should expect a response from the server
func (msg *BasicGet) Wait() bool {
	return true
}

// Write serializes this message to the provided writer
func (msg *BasicGet) Write(w io.Writer) (err error) {
	var bits byte

	if err = binary.Write(w, binary.BigEndian, msg.reserved1); err != nil {
		return
	}

	if err = writeShortstr(w, msg.Queue); err != nil {
		return
	}

	if msg.NoAck {
		bits |= 1 << 0
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	return
}

// Read deserializes this message from the provided reader
func (msg *BasicGet) Read(r io.Reader) (err error) {
	var bits byte

	if err = binary.Read(r, binary.BigEndian, &msg.reserved1); err != nil {
		return
	}

	if msg.Queue, err = readShortstr(r); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.NoAck = (bits&(1<<0) > 0)

	return
}

// BasicGetOk represents the AMQP message basic.get-ok
type BasicGetOk struct {
	DeliveryTag  uint64
	Redelivered  bool
	Exchange     string
	RoutingKey   string
	MessageCount uint32
	Properties   Properties
	Body         []byte
}

// ID returns the AMQP class and method identifiers for this message
func (msg *BasicGetOk) ID() (uint16, uint16) {
	return 60, 71
}

// Wait returns true when the client should expect a response from the server
func (msg *BasicGetOk) Wait() bool {
	return true
}

// GetContent returns the Properties and Body from this message
func (msg *BasicGetOk) GetContent() (Properties, []byte) {
	return msg.Properties, msg.Body
}

// SetContent sets the Properties and Body for serialization
func (msg *BasicGetOk) SetContent(props Properties, body []byte) {
	msg.Properties, msg.Body = props, body
}

// Write serializes this message to the provided writer
func (msg *BasicGetOk) Write(w io.Writer) (err error) {
	var bits byte

	if err = binary.Write(w, binary.BigEndian, msg.DeliveryTag); err != nil {
		return
	}

	if msg.Redelivered {
		bits |= 1 << 0
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	if err = writeShortstr(w, msg.Exchange); err != nil {
		return
	}
	if err = writeShortstr(w, msg.RoutingKey); err != nil {
		return
	}

	if err = binary.Write(w, binary.BigEndian, msg.MessageCount); err != nil {
		return
	}

	return
}

// Read deserializes this message from the provided reader
func (msg *BasicGetOk) Read(r io.Reader) (err error) {
	var bits byte

	if err = binary.Read(r, binary.BigEndian, &msg.DeliveryTag); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.Redelivered = (bits&(1<<0) > 0)

	if msg.Exchange, err = readShortstr(r); err != nil {
		return
	}
	if msg.RoutingKey, err = readShortstr(r); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &msg.MessageCount); err != nil {
		return
	}

	return
}

// BasicGetEmpty represents the AMQP message basic.get-empty
type BasicGetEmpty struct {
	reserved1 string
}

// ID returns the AMQP class and method identifiers for this message
func (msg *BasicGetEmpty) ID() (uint16, uint16) {
	return 60, 72
}

// Wait returns true when the client should expect a response from the server
func (msg *BasicGetEmpty) Wait() bool {
	return true
}

// Write serializes this message to the provided writer
func (msg *BasicGetEmpty) Write(w io.Writer) (err error) {

	if err = writeShortstr(w, msg.reserved1); err != nil {
		return
	}

	return
}

// Read deserializes this message from the provided reader
func (msg *BasicGetEmpty) Read(r io.Reader) (err error) {

	if msg.reserved1, err = readShortstr(r); err != nil {
		return
	}

	return
}

// BasicAck represents the AMQP message basic.ack
type BasicAck struct {
	DeliveryTag uint64
	Multiple    bool
}

// ID returns the AMQP class and method identifiers for this message
func (msg *BasicAck) ID() (uint16, uint16) {
	return 60, 80
}

// Wait returns true when the client should expect a response from the server
func (msg *BasicAck) Wait() bool {
	return false
}

// Write serializes this message to the provided writer
func (msg *BasicAck) Write(w io.Writer) (err error) {
	var bits byte

	if err = binary.Write(w, binary.BigEndian, msg.DeliveryTag); err != nil {
		return
	}

	if msg.Multiple {
		bits |= 1 << 0
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	return
}

// Read deserializes this message from the provided reader
func (msg *BasicAck) Read(r io.Reader) (err error) {
	var bits byte

	if err = binary.Read(r, binary.BigEndian, &msg.DeliveryTag); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.Multiple = (bits&(1<<0) > 0)

	return
}

// BasicReject represents the AMQP message basic.reject
type BasicReject struct {
	DeliveryTag uint64
	Requeue     bool
}

// ID returns the AMQP class and method identifiers for this message
func (msg *BasicReject) ID() (uint16, uint16) {
	return 60, 90
}

// Wait returns true when the client should expect a response from the server
func (msg *BasicReject) Wait() bool {
	return false
}

// Write serializes this message to the provided writer
func (msg *BasicReject) Write(w io.Writer) (err error) {
	var bits byte

	if err = binary.Write(w, binary.BigEndian, msg.DeliveryTag); err != nil {
		return
	}

	if msg.Requeue {
		bits |= 1 << 0
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	return
}

// Read deserializes this message from the provided reader
func (msg *BasicReject) Read(r io.Reader) (err error) {
	var bits byte

	if err = binary.Read(r, binary.BigEndian, &msg.DeliveryTag); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.Requeue = (bits&(1<<0) > 0)

	return
}

// BasicRecoverAsync represents the AMQP message basic.recover-async
type BasicRecoverAsync struct {
	Requeue bool
}

// ID returns the AMQP class and method identifiers for this message
func (msg *BasicRecoverAsync) ID() (uint16, uint16) {
	return 60, 100
}

// Wait returns true when the client should expect a response from the server
func (msg *BasicRecoverAsync) Wait() bool {
	return false
}

// Write serializes this message to the provided writer
func (msg *BasicRecoverAsync) Write(w io.Writer) (err error) {
	var bits byte

	if msg.Requeue {
		bits |= 1 << 0
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	return
}

// Read deserializes this message from the provided reader
func (msg *BasicRecoverAsync) Read(r io.Reader) (err error) {
	var bits byte

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.Requeue = (bits&(1<<0) > 0)

	return
}

// BasicRecover represents the AMQP message basic.recover
type BasicRecover struct {
	Requeue bool
}

// ID returns the AMQP class and method identifiers for this message
func (msg *BasicRecover) ID() (uint16, uint16) {
	return 60, 110
}

// Wait returns true when the client should expect a response from the server
func (msg *BasicRecover) Wait() bool {
	return true
}

// Write serializes this message to the provided writer
func (msg *BasicRecover) Write(w io.Writer) (err error) {
	var bits byte

	if msg.Requeue {
		bits |= 1 << 0
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	return
}

// Read deserializes this message from the provided reader
func (msg *BasicRecover) Read(r io.Reader) (err error) {
	var bits byte

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.Requeue = (bits&(1<<0) > 0)

	return
}

// BasicRecoverOk represents the AMQP message basic.recover-ok
type BasicRecoverOk struct {
}

// ID returns the AMQP class and method identifiers for this message
func (msg *BasicRecoverOk) ID() (uint16, uint16) {
	return 60, 111
}

// Wait returns true when the client should expect a response from the server
func (msg *BasicRecoverOk) Wait() bool {
	return true
}

// Write serializes this message to the provided writer
func (msg *BasicRecoverOk) Write(w io.Writer) (err error) {

	return
}

// Read deserializes this message from the provided reader
func (msg *BasicRecoverOk) Read(r io.Reader) (err error) {

	return
}

// BasicNack represents the AMQP message basic.nack
type BasicNack struct {
	DeliveryTag uint64
	Multiple    bool
	Requeue     bool
}

// ID returns the AMQP class and method identifiers for this message
func (msg *BasicNack) ID() (uint16, uint16) {
	return 60, 120
}

// Wait returns true when the client should expect a response from the server
func (msg *BasicNack) Wait() bool {
	return false
}

// Write serializes this message to the provided writer
func (msg *BasicNack) Write(w io.Writer) (err error) {
	var bits byte

	if err = binary.Write(w, binary.BigEndian, msg.DeliveryTag); err != nil {
		return
	}

	if msg.Multiple {
		bits |= 1 << 0
	}

	if msg.Requeue {
		bits |= 1 << 1
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	return
}

// Read deserializes this message from the provided reader
func (msg *BasicNack) Read(r io.Reader) (err error) {
	var bits byte

	if err = binary.Read(r, binary.BigEndian, &msg.DeliveryTag); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.Multiple = (bits&(1<<0) > 0)
	msg.Requeue = (bits&(1<<1) > 0)

	return
}

// TxSelect represents the AMQP message tx.select
type TxSelect struct {
}

// ID returns the AMQP class and method identifiers for this message
func (msg *TxSelect) ID() (uint16, uint16) {
	return 90, 10
}

// Wait returns true when the client should expect a response from the server
func (msg *TxSelect) Wait() bool {
	return true
}

// Write serializes this message to the provided writer
func (msg *TxSelect) Write(w io.Writer) (err error) {

	return
}

// Read deserializes this message from the provided reader
func (msg *TxSelect) Read(r io.Reader) (err error) {

	return
}

// TxSelectOk represents the AMQP message tx.select-ok
type TxSelectOk struct {
}

// ID returns the AMQP class and method identifiers for this message
func (msg *TxSelectOk) ID() (uint16, uint16) {
	return 90, 11
}

// Wait returns true when the client should expect a response from the server
func (msg *TxSelectOk) Wait() bool {
	return true
}

// Write serializes this message to the provided writer
func (msg *TxSelectOk) Write(w io.Writer) (err error) {

	return
}

// Read deserializes this message from the provided reader
func (msg *TxSelectOk) Read(r io.Reader) (err error) {

	return
}

// TxCommit represents the AMQP message tx.commit
type TxCommit struct {
}

// ID returns the AMQP class and method identifiers for this message
func (msg *TxCommit) ID() (uint16, uint16) {
	return 90, 20
}

// Wait returns true when the client should expect a response from the server
func (msg *TxCommit) Wait() bool {
	return true
}

// Write serializes this message to the provided writer
func (msg *TxCommit) Write(w io.Writer) (err error) {

	return
}

// Read deserializes this message from the provided reader
func (msg *TxCommit) Read(r io.Reader) (err error) {

	return
}

// TxCommitOk represents the AMQP message tx.commit-ok
type TxCommitOk struct {
}

// ID returns the AMQP class and method identifiers for this message
func (msg *TxCommitOk) ID() (uint16, uint16) {
	return 90, 21
}

// Wait returns true when the client should expect a response from the server
func (msg *TxCommitOk) Wait() bool {
	return true
}

// Write serializes this message to the provided writer
func (msg *TxCommitOk) Write(w io.Writer) (err error) {

	return
}

// Read deserializes this message from the provided reader
func (msg *TxCommitOk) Read(r io.Reader) (err error) {

	return
}

// TxRollback represents the AMQP message tx.rollback
type TxRollback struct {
}

// ID returns the AMQP class and method identifiers for this message
func (msg *TxRollback) ID() (uint16, uint16) {
	return 90, 30
}

// Wait returns true when the client should expect a response from the server
func (msg *TxRollback) Wait() bool {
	return true
}

// Write serializes this message to the provided writer
func (msg *TxRollback) Write(w io.Writer) (err error) {

	return
}

// Read deserializes this message from the provided reader
func (msg *TxRollback) Read(r io.Reader) (err error) {

	return
}

// TxRollbackOk represents the AMQP message tx.rollback-ok
type TxRollbackOk struct {
}

// ID returns the AMQP class and method identifiers for this message
func (msg *TxRollbackOk) ID() (uint16, uint16) {
	return 90, 31
}

// Wait returns true when the client should expect a response from the server
func (msg *TxRollbackOk) Wait() bool {
	return true
}

// Write serializes this message to the provided writer
func (msg *TxRollbackOk) Write(w io.Writer) (err error) {

	return
}

// Read deserializes this message from the provided reader
func (msg *TxRollbackOk) Read(r io.Reader) (err error) {

	return
}

// ConfirmSelect represents the AMQP message confirm.select
type ConfirmSelect struct {
	Nowait bool
}

// ID returns the AMQP class and method identifiers for this message
func (msg *ConfirmSelect) ID() (uint16, uint16) {
	return 85, 10
}

// Wait returns true when the client should expect a response from the server
func (msg *ConfirmSelect) Wait() bool {
	return true
}

// Write serializes this message to the provided writer
func (msg *ConfirmSelect) Write(w io.Writer) (err error) {
	var bits byte

	if msg.Nowait {
		bits |= 1 << 0
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	return
}

// Read deserializes this message from the provided reader
func (msg *ConfirmSelect) Read(r io.Reader) (err error) {
	var bits byte

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.Nowait = (bits&(1<<0) > 0)

	return
}

// ConfirmSelectOk represents the AMQP message confirm.select-ok
type ConfirmSelectOk struct {
}

// ID returns the AMQP class and method identifiers for this message
func (msg *ConfirmSelectOk) ID() (uint16, uint16) {
	return 85, 11
}

// Wait returns true when the client should expect a response from the server
func (msg *ConfirmSelectOk) Wait() bool {
	return true
}

// Write serializes this message to the provided writer
func (msg *ConfirmSelectOk) Write(w io.Writer) (err error) {

	return
}

// Read deserializes this message from the provided reader
func (msg *ConfirmSelectOk) Read(r io.Reader) (err error) {

	return
}

func (r *Reader) parseMethodFrame(channel uint16, size uint32) (f Frame, err error) {
	mf := &MethodFrame{
		ChannelId: channel,
	}

	if err = binary.Read(r.r, binary.BigEndian, &mf.ClassId); err != nil {
		return
	}

	if err = binary.Read(r.r, binary.BigEndian, &mf.MethodId); err != nil {
		return
	}

	switch mf.ClassId {

	case 10: // connection
		switch mf.MethodId {

		case 10: // connection start
			//fmt.Println("NextMethod: class:10 method:10")
			method := &ConnectionStart{}
			if err = method.Read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 11: // connection start-ok
			//fmt.Println("NextMethod: class:10 method:11")
			method := &ConnectionStartOk{}
			if err = method.Read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 20: // connection secure
			//fmt.Println("NextMethod: class:10 method:20")
			method := &ConnectionSecure{}
			if err = method.Read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 21: // connection secure-ok
			//fmt.Println("NextMethod: class:10 method:21")
			method := &ConnectionSecureOk{}
			if err = method.Read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 30: // connection tune
			//fmt.Println("NextMethod: class:10 method:30")
			method := &ConnectionTune{}
			if err = method.Read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 31: // connection tune-ok
			//fmt.Println("NextMethod: class:10 method:31")
			method := &ConnectionTuneOk{}
			if err = method.Read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 40: // connection open
			//fmt.Println("NextMethod: class:10 method:40")
			method := &ConnectionOpen{}
			if err = method.Read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 41: // connection open-ok
			//fmt.Println("NextMethod: class:10 method:41")
			method := &ConnectionOpenOk{}
			if err = method.Read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 50: // connection close
			//fmt.Println("NextMethod: class:10 method:50")
			method := &ConnectionClose{}
			if err = method.Read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 51: // connection close-ok
			//fmt.Println("NextMethod: class:10 method:51")
			method := &ConnectionCloseOk{}
			if err = method.Read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 60: // connection blocked
			//fmt.Println("NextMethod: class:10 method:60")
			method := &ConnectionBlocked{}
			if err = method.Read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 61: // connection unblocked
			//fmt.Println("NextMethod: class:10 method:61")
			method := &ConnectionUnblocked{}
			if err = method.Read(r.r); err != nil {
				return
			}
			mf.Method = method

		default:
			return nil, fmt.Errorf("Bad method frame, unknown method %d for class %d", mf.MethodId, mf.ClassId)
		}

	case 20: // channel
		switch mf.MethodId {

		case 10: // channel open
			//fmt.Println("NextMethod: class:20 method:10")
			method := &ChannelOpen{}
			if err = method.Read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 11: // channel open-ok
			//fmt.Println("NextMethod: class:20 method:11")
			method := &ChannelOpenOk{}
			if err = method.Read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 20: // channel flow
			//fmt.Println("NextMethod: class:20 method:20")
			method := &ChannelFlow{}
			if err = method.Read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 21: // channel flow-ok
			//fmt.Println("NextMethod: class:20 method:21")
			method := &ChannelFlowOk{}
			if err = method.Read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 40: // channel close
			//fmt.Println("NextMethod: class:20 method:40")
			method := &ChannelClose{}
			if err = method.Read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 41: // channel close-ok
			//fmt.Println("NextMethod: class:20 method:41")
			method := &ChannelCloseOk{}
			if err = method.Read(r.r); err != nil {
				return
			}
			mf.Method = method

		default:
			return nil, fmt.Errorf("Bad method frame, unknown method %d for class %d", mf.MethodId, mf.ClassId)
		}

	case 40: // exchange
		switch mf.MethodId {

		case 10: // exchange declare
			//fmt.Println("NextMethod: class:40 method:10")
			method := &ExchangeDeclare{}
			if err = method.Read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 11: // exchange declare-ok
			//fmt.Println("NextMethod: class:40 method:11")
			method := &ExchangeDeclareOk{}
			if err = method.Read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 20: // exchange delete
			//fmt.Println("NextMethod: class:40 method:20")
			method := &ExchangeDelete{}
			if err = method.Read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 21: // exchange delete-ok
			//fmt.Println("NextMethod: class:40 method:21")
			method := &ExchangeDeleteOk{}
			if err = method.Read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 30: // exchange bind
			//fmt.Println("NextMethod: class:40 method:30")
			method := &ExchangeBind{}
			if err = method.Read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 31: // exchange bind-ok
			//fmt.Println("NextMethod: class:40 method:31")
			method := &ExchangeBindOk{}
			if err = method.Read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 40: // exchange unbind
			//fmt.Println("NextMethod: class:40 method:40")
			method := &ExchangeUnbind{}
			if err = method.Read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 51: // exchange unbind-ok
			//fmt.Println("NextMethod: class:40 method:51")
			method := &ExchangeUnbindOk{}
			if err = method.Read(r.r); err != nil {
				return
			}
			mf.Method = method

		default:
			return nil, fmt.Errorf("Bad method frame, unknown method %d for class %d", mf.MethodId, mf.ClassId)
		}

	case 50: // queue
		switch mf.MethodId {

		case 10: // queue declare
			//fmt.Println("NextMethod: class:50 method:10")
			method := &QueueDeclare{}
			if err = method.Read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 11: // queue declare-ok
			//fmt.Println("NextMethod: class:50 method:11")
			method := &QueueDeclareOk{}
			if err = method.Read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 20: // queue bind
			//fmt.Println("NextMethod: class:50 method:20")
			method := &QueueBind{}
			if err = method.Read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 21: // queue bind-ok
			//fmt.Println("NextMethod: class:50 method:21")
			method := &QueueBindOk{}
			if err = method.Read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 50: // queue unbind
			//fmt.Println("NextMethod: class:50 method:50")
			method := &QueueUnbind{}
			if err = method.Read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 51: // queue unbind-ok
			//fmt.Println("NextMethod: class:50 method:51")
			method := &QueueUnbindOk{}
			if err = method.Read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 30: // queue purge
			//fmt.Println("NextMethod: class:50 method:30")
			method := &QueuePurge{}
			if err = method.Read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 31: // queue purge-ok
			//fmt.Println("NextMethod: class:50 method:31")
			method := &QueuePurgeOk{}
			if err = method.Read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 40: // queue delete
			//fmt.Println("NextMethod: class:50 method:40")
			method := &QueueDelete{}
			if err = method.Read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 41: // queue delete-ok
			//fmt.Println("NextMethod: class:50 method:41")
			method := &QueueDeleteOk{}
			if err = method.Read(r.r); err != nil {
				return
			}
			mf.Method = method

		default:
			return nil, fmt.Errorf("Bad method frame, unknown method %d for class %d", mf.MethodId, mf.ClassId)
		}

	case 60: // basic
		switch mf.MethodId {

		case 10: // basic qos
			//fmt.Println("NextMethod: class:60 method:10")
			method := &BasicQos{}
			if err = method.Read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 11: // basic qos-ok
			//fmt.Println("NextMethod: class:60 method:11")
			method := &BasicQosOk{}
			if err = method.Read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 20: // basic consume
			//fmt.Println("NextMethod: class:60 method:20")
			method := &BasicConsume{}
			if err = method.Read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 21: // basic consume-ok
			//fmt.Println("NextMethod: class:60 method:21")
			method := &BasicConsumeOk{}
			if err = method.Read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 30: // basic cancel
			//fmt.Println("NextMethod: class:60 method:30")
			method := &BasicCancel{}
			if err = method.Read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 31: // basic cancel-ok
			//fmt.Println("NextMethod: class:60 method:31")
			method := &BasicCancelOk{}
			if err = method.Read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 40: // basic publish
			//fmt.Println("NextMethod: class:60 method:40")
			method := &BasicPublish{}
			if err = method.Read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 50: // basic return
			//fmt.Println("NextMethod: class:60 method:50")
			method := &BasicReturn{}
			if err = method.Read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 60: // basic deliver
			//fmt.Println("NextMethod: class:60 method:60")
			method := &BasicDeliver{}
			if err = method.Read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 70: // basic get
			//fmt.Println("NextMethod: class:60 method:70")
			method := &BasicGet{}
			if err = method.Read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 71: // basic get-ok
			//fmt.Println("NextMethod: class:60 method:71")
			method := &BasicGetOk{}
			if err = method.Read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 72: // basic get-empty
			//fmt.Println("NextMethod: class:60 method:72")
			method := &BasicGetEmpty{}
			if err = method.Read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 80: // basic ack
			//fmt.Println("NextMethod: class:60 method:80")
			method := &BasicAck{}
			if err = method.Read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 90: // basic reject
			//fmt.Println("NextMethod: class:60 method:90")
			method := &BasicReject{}
			if err = method.Read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 100: // basic recover-async
			//fmt.Println("NextMethod: class:60 method:100")
			method := &BasicRecoverAsync{}
			if err = method.Read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 110: // basic recover
			//fmt.Println("NextMethod: class:60 method:110")
			method := &BasicRecover{}
			if err = method.Read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 111: // basic recover-ok
			//fmt.Println("NextMethod: class:60 method:111")
			method := &BasicRecoverOk{}
			if err = method.Read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 120: // basic nack
			//fmt.Println("NextMethod: class:60 method:120")
			method := &BasicNack{}
			if err = method.Read(r.r); err != nil {
				return
			}
			mf.Method = method

		default:
			return nil, fmt.Errorf("Bad method frame, unknown method %d for class %d", mf.MethodId, mf.ClassId)
		}

	case 90: // tx
		switch mf.MethodId {

		case 10: // tx select
			//fmt.Println("NextMethod: class:90 method:10")
			method := &TxSelect{}
			if err = method.Read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 11: // tx select-ok
			//fmt.Println("NextMethod: class:90 method:11")
			method := &TxSelectOk{}
			if err = method.Read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 20: // tx commit
			//fmt.Println("NextMethod: class:90 method:20")
			method := &TxCommit{}
			if err = method.Read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 21: // tx commit-ok
			//fmt.Println("NextMethod: class:90 method:21")
			method := &TxCommitOk{}
			if err = method.Read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 30: // tx rollback
			//fmt.Println("NextMethod: class:90 method:30")
			method := &TxRollback{}
			if err = method.Read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 31: // tx rollback-ok
			//fmt.Println("NextMethod: class:90 method:31")
			method := &TxRollbackOk{}
			if err = method.Read(r.r); err != nil {
				return
			}
			mf.Method = method

		default:
			return nil, fmt.Errorf("Bad method frame, unknown method %d for class %d", mf.MethodId, mf.ClassId)
		}

	case 85: // confirm
		switch mf.MethodId {

		case 10: // confirm select
			//fmt.Println("NextMethod: class:85 method:10")
			method := &ConfirmSelect{}
			if err = method.Read(r.r); err != nil {
				return
			}
			mf.Method = method

		case 11: // confirm select-ok
			//fmt.Println("NextMethod: class:85 method:11")
			method := &ConfirmSelectOk{}
			if err = method.Read(r.r); err != nil {
				return
			}
			mf.Method = method

		default:
			return nil, fmt.Errorf("Bad method frame, unknown method %d for class %d", mf.MethodId, mf.ClassId)
		}

	default:
		return nil, fmt.Errorf("Bad method frame, unknown class %d", mf.ClassId)
	}

	return mf, nil
}
