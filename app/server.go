package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"time"
)

func exitWithError(err error) {
	fmt.Println(err.Error())
	os.Exit(1)
}

type TimeValue struct {
	value     string
	createdAt time.Time
	ex        int
}

func (v TimeValue) isExpired() bool {
	if v.ex == 0 {
		return false
	}
	expiry := v.createdAt.Add(time.Millisecond * time.Duration(v.ex))
	return expiry.Before(time.Now())

}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		exitWithError(fmt.Errorf("failed to bind to port 6379: %v", err))
	}

	// keyTable := make(map[string]string)
	expiryTable := make(map[string]TimeValue)

	for {
		conn, err := l.Accept()
		if err != nil {
			exitWithError(fmt.Errorf("Error accepting connection: %v", err))
		}

		go handleConnection(conn, expiryTable)
	}
}

func handleConnection(conn net.Conn, expiryTable map[string]TimeValue) {
	defer conn.Close()

	for {
		reader := bufio.NewReaderSize(conn, 2056)

		val, err := parseRESP(reader)
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			exitWithError(err)
		}

		command := val.Array[0].String
		args := val.Array[1:]
		fmt.Println("command: ", command)
		fmt.Println("args: ", args)

		switch command {
		case "ping":
			conn.Write([]byte(fmt.Sprintf("+%s\r\n", "PONG")))
		case "echo":
			conn.Write([]byte(fmt.Sprintf("$%d\r\n%s\r\n", len(args[0].String), args[0].String)))
		case "set":
			handleSet(args, expiryTable, conn)
		case "get":
			handleGet(args, expiryTable, conn)
		default:
			conn.Write([]byte(fmt.Sprintf("-Error unknown command %s\r\n", command)))
		}
	}
}

func handleSet(args []Val, expiryTable map[string]TimeValue, conn net.Conn) {
	if len(args) < 2 {
		return
	}

	key := args[0].String
	value := TimeValue{args[1].String, time.Now(), 0}
	if len(args) == 4 && args[2].String == "px" {
		ms, err := strconv.Atoi(args[3].String)
		if err != nil {
			fmt.Println("Failed to parse. %s", ms)
			return
		}
		value.ex = ms
	}

	expiryTable[key] = value

	conn.Write([]byte("+OK\r\n"))
}

func handleGet(args []Val, expiryTable map[string]TimeValue, conn net.Conn) {
	if len(args) != 1 {
		return
	}
	value, ok := expiryTable[args[0].String]
	if ok && !value.isExpired() {
		conn.Write([]byte(fmt.Sprintf("$%d\r\n%s\r\n", len(value.value), value.value)))
	} else {
		conn.Write([]byte("$-1\r\n"))
	}
}

const (
	SimpleString byte = '+'
	Error        byte = '-'
	Integer      byte = ':'
	BulkString   byte = '$'
	Array        byte = '*'
)

type Val struct {
	Type   byte
	String string
	Array  []Val
}

func parseVal(val Val) []byte {
	switch val.Type {
	case SimpleString:
		return []byte(fmt.Sprintf("+%s\r\n", val.String))
	case BulkString:
		return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(val.String), val.String))
	case Array:
		for _, ele := range val.Array {
			g := parseVal(ele)
			if g != nil {
				return g
			}
		}
	default:
		return nil
	}

	return nil
}

func parseRESP(b *bufio.Reader) (Val, error) {
	firstByte, err := b.ReadByte()
	if err != nil {
		return Val{}, err
	}
	switch firstByte {
	case SimpleString:
		return parseSimpleString(b)

	case BulkString:
		return parseBulkString(b)

	case Array:
		return parseArray(b)

	default:
		return Val{}, fmt.Errorf("Invalid RESP request type: %s", string(firstByte))

	}
}

func parseArray(b *bufio.Reader) (Val, error) {
	sizeByte, err := b.ReadBytes('\n')
	if err != nil {
		exitWithError(err)
	}
	sizeByte = bytes.TrimSpace(sizeByte)
	size, err := strconv.Atoi(string(sizeByte))
	if err != nil {
		exitWithError(err)
	}

	result := make([]Val, 0)
	for i := 0; i < size; i++ {
		ele, err := parseRESP(b)
		if err != nil {
			return Val{}, err
		}

		result = append(result, ele)
	}

	return Val{
		Type:  Array,
		Array: result,
	}, nil
}

func parseSimpleString(b *bufio.Reader) (Val, error) {
	byteStream, err := b.ReadBytes('\n')
	if err != nil {
		exitWithError(err)
	}

	return Val{
		Type:   SimpleString,
		String: string(bytes.TrimSpace(byteStream)),
	}, nil
}

func parseBulkString(b *bufio.Reader) (Val, error) {
	sizeByte, err := b.ReadBytes('\n')
	if err != nil {
		exitWithError(err)
	}
	sizeByte = bytes.TrimSpace(sizeByte)
	bulkStringByte, err := b.ReadBytes('\n')
	if err != nil {
		return Val{}, fmt.Errorf("Failed to convert BulkString: %s", string(bulkStringByte))
	}

	bulkString := string(bytes.TrimSpace(bulkStringByte))

	return Val{
		Type:   BulkString,
		String: bulkString,
	}, nil
}
