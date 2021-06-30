package session

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
	"net"
	"os"

	"weavelab.xyz/ethr/lib"
)

func CreateAckMsg() (msg *lib.Msg) {
	msg = &lib.Msg{Version: 0, Type: lib.Ack}
	msg.Ack = &lib.MsgAck{}
	return
}

func CreateSynMsg(testID lib.TestID, clientParam lib.ClientParams) (msg *lib.Msg) {
	msg = &lib.Msg{Version: 0, Type: lib.Syn}
	msg.Syn = &lib.MsgSyn{}
	msg.Syn.TestID = testID
	msg.Syn.ClientParam = clientParam
	return
}

func (s *Session) HandshakeWithServer(test *Test, conn net.Conn) error {
	msg := CreateSynMsg(test.ID, test.ClientParam)
	err := s.Send(conn, msg)
	if err != nil {
		return fmt.Errorf("failed to send SYN message: %w", err)
	}
	resp, err := s.Receive(conn)
	if err != nil {
		return err
	}
	if resp.Type != lib.Ack {
		return fmt.Errorf("failed to receive ACK message: %w", os.ErrInvalid)
	}
	return nil
}

func (s *Session) HandshakeWithClient(conn net.Conn) (testID lib.TestID, clientParam lib.ClientParams, err error) {
	msg, err := s.Receive(conn)
	if err != nil {
		return
	}
	if msg.Type != lib.Syn {
		err = os.ErrInvalid
		return
	}
	testID = msg.Syn.TestID
	clientParam = msg.Syn.ClientParam
	ack := CreateAckMsg()
	err = s.Send(conn, ack)
	return
}

func (s *Session) Receive(conn net.Conn) (msg *lib.Msg, err error) {
	msg = &lib.Msg{}
	msg.Type = lib.Inv
	msgBytes := make([]byte, 4)
	_, err = io.ReadFull(conn, msgBytes)
	if err != nil {
		//Logger.Debug("Error receiving message on control channel. Error: %v", err)
		return
	}
	msgSize := binary.BigEndian.Uint32(msgBytes[0:])
	// Max ethr message size as 16K sent over gob.
	if msgSize > 16384 {
		return
	}
	msgBytes = make([]byte, msgSize)
	_, err = io.ReadFull(conn, msgBytes)
	if err != nil {
		//Logger.Debug("Error receiving message on control channel. Error: %v", err)
		return
	}
	msg = decodeMsg(msgBytes)
	return
}

func (s *Session) ReceiveFromBuffer(msgBytes []byte) (msg *lib.Msg) {
	msg = decodeMsg(msgBytes)
	return
}

func (s *Session) Send(conn net.Conn, msg *lib.Msg) (err error) {
	msgBytes, err := encodeMsg(msg)
	if err != nil {
		Logger.Debug("Error sending message on control channel. Message: %v, Error: %v", msg, err)
		return
	}
	msgSize := len(msgBytes)
	tempBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(tempBuf[0:], uint32(msgSize))
	_, err = conn.Write(tempBuf)
	if err != nil {
		Logger.Debug("Error sending message on control channel. Message: %v, Error: %v", msg, err)
	}
	_, err = conn.Write(msgBytes)
	if err != nil {
		Logger.Debug("Error sending message on control channel. Message: %v, Error: %v", msg, err)
	}
	return err
}

func decodeMsg(msgBytes []byte) (msg *lib.Msg) {
	msg = &lib.Msg{}
	buffer := bytes.NewBuffer(msgBytes)
	decoder := gob.NewDecoder(buffer)
	err := decoder.Decode(msg)
	if err != nil {
		Logger.Debug("Failed to decode message using Gob: %v", err)
		msg.Type = lib.Inv
	}
	return
}

func encodeMsg(msg *lib.Msg) (msgBytes []byte, err error) {
	var writeBuffer bytes.Buffer
	encoder := gob.NewEncoder(&writeBuffer)
	err = encoder.Encode(msg)
	if err != nil {
		Logger.Debug("Failed to encode message using Gob: %v", err)
		return
	}
	msgBytes = writeBuffer.Bytes()
	return
}
