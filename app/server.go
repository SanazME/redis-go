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
)

func exitWithError(err error) {
	fmt.Println(err.Error())
	os.Exit(1)
}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		exitWithError(fmt.Errorf("failed to bind to port 6379: %v", err))
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			exitWithError(fmt.Errorf("Error accepting connection: %v", err))
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	for {
		reader := bufio.NewReaderSize(conn, 2056)

		val, err := parseRESP(reader)
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			exitWithError(err)
		}

		_, err = conn.Write(parseVal(val))
		if err != nil {
			exitWithError(fmt.Errorf("Error writing to conneciton %v", err))
		}
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
	case Integer:
	case BulkString:
		return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(val.String), val.String))
		// for zero string
		// for nil : $-1\r\n
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
		fmt.Println("Simple string")

	case Error:
		fmt.Println("Error")

	case Integer:
		fmt.Println("Integer")

	case BulkString:
		return parseBulkString(b)

	case Array:
		return parseArray(b)

	default:
		return Val{}, fmt.Errorf("Invalid RESP request type: %s", string(firstByte))

	}
	return Val{}, nil
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

	if bulkString == "ping" {
		return Val{
			Type:   SimpleString,
			String: "PONG",
		}, nil
	}
	if bulkString == "echo" {
		return Val{}, nil
	}

	return Val{
		Type:   BulkString,
		String: bulkString,
	}, nil
}
