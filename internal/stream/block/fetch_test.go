package block

import (
	"bytes"
	"io"
	"log"
	"net"

	"testing"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/lomocoin/lws/internal/coreclient"
	"github.com/lomocoin/lws/internal/coreclient/DBPMsg/go/dbp"
	"github.com/lomocoin/lws/internal/coreclient/DBPMsg/go/lws"
	cclientModule "github.com/lomocoin/lws/internal/stream/cclient"
	"github.com/lomocoin/lws/test/helper"
)

// func TestFetch500(t *testing.T) {
// 	b := &BlockFetcher{}

// 	cclient.StartCoreClient()

// 	hash, _ := hex.DecodeString("029477812d10f0c8c3a9be59dce1203dedc6f8cf3b6d8f8f9f4d981cd01495b9")
// 	b.fetch(hash, 500)
// }

func TestFetchForkedChain(t *testing.T) {
	helper.ResetDb()
	helper.LoadTestSeed("seedRecovery.sql")

	serverConn, clientConn := net.Pipe()

	go (func(conn io.ReadWriteCloser) {
		var err error
		var hash []byte
		var results []*lws.Block
		mockServer := coreclient.NewMockServer(conn)

		// 1 round - fetch
		if err := mockServer.Decoder.ReadMsg(&mockServer.WireRes); err != nil {
			t.Fatalf("ReadMsg failed[%s]", err)
		}

		method, ok := mockServer.WireRes.Response.(*dbp.Method)
		if !ok {
			t.Fatalf("received non-method message type:[%s]", mockServer.WireRes.MsgType)
		}

		blockArg := &lws.GetBlocksArg{}
		err = ptypes.UnmarshalAny(method.Params, blockArg)
		if err != nil {
			t.Fatalf("received non-getblocks message")
		}

		hash = []byte("0000000000000000000000000005")
		if bytes.Compare(blockArg.Hash, hash) != 0 {
			t.Fatalf("first fetch hash expect [%s], but [%s]", hash, blockArg.Hash)
		}

		results = []*lws.Block{
			&lws.Block{
				NHeight:  4,
				Hash:     []byte("0000000000000000000000000005"),
				HashPrev: []byte("0000000000000000000000000004"),
			},
			&lws.Block{
				NHeight:  5,
				Hash:     []byte("1000000000000000000000000006"),
				HashPrev: []byte("1000000000000000000000000005"),
			},
			&lws.Block{
				NHeight:  6,
				Hash:     []byte("1000000000000000000000000007"),
				HashPrev: []byte("1000000000000000000000000006"),
			},
		}

		serializedResults := make([]*any.Any, len(results))
		for idx, block := range results {
			anyAny, err := ptypes.MarshalAny(block)
			if err != nil {
				log.Fatal("could not serialize any field")
			}

			serializedResults[idx] = anyAny
		}

		resultMsg := &dbp.Result{
			Error:  "",
			Result: serializedResults,
		}

		log.Print("result Msg")

		mockServer.WireReq.ID = method.Id
		mockServer.WireReq.Request = resultMsg
		if err := mockServer.Encoder.WriteMsg(&mockServer.WireReq); err != nil {
			t.Fatalf("WriteMsg failed[%s]", err)
		}
		if err := mockServer.Encoder.Flush(); err != nil {
			t.Fatalf("Write flush failed: [%s]", err)
		}

		log.Print("write 1 response done")

		// 2 round - fetch
		if err := mockServer.Decoder.ReadMsg(&mockServer.WireRes); err != nil {
			t.Fatalf("ReadMsg failed[%s]", err)
		}

		method, ok = mockServer.WireRes.Response.(*dbp.Method)
		if !ok {
			t.Fatalf("received non-method message type:[%s]", mockServer.WireRes.MsgType)
		}

		blockArg = &lws.GetBlocksArg{}
		err = ptypes.UnmarshalAny(method.Params, blockArg)
		if err != nil {
			t.Fatalf("received non-getblocks message")
		}

		log.Print("2 round- block args get")

		hash = []byte("1000000000000000000000000005")
		log.Print("hash 1", hash)
		log.Print("hash 2", blockArg.Hash)
		if bytes.Compare(blockArg.Hash, hash) != 0 {
			log.Print("hello？")
			t.Fatalf("2 round fetch hash expect [%s], but [%s]", hash, blockArg.Hash)
		}

		if blockArg.Number != 1 {
			log.Print("hello Number？")
			t.Fatalf("2 round fetch number expect 1, but [%d]", blockArg.Number)
		}

		log.Print("2 round- results")
		results = []*lws.Block{
			&lws.Block{
				NHeight:  4,
				Hash:     []byte("1000000000000000000000000005"),
				HashPrev: []byte("1000000000000000000000000004"),
			},
		}

		serializedResults = make([]*any.Any, len(results))
		for idx, block := range results {
			anyAny, err := ptypes.MarshalAny(block)
			if err != nil {
				log.Fatal("could not serialize any field")
			}

			serializedResults[idx] = anyAny
		}

		resultMsg = &dbp.Result{
			Error:  "",
			Result: serializedResults,
		}
		log.Print("2 round- resultMsg")

		mockServer.WireReq.ID = method.Id
		mockServer.WireReq.Request = resultMsg
		if err := mockServer.Encoder.WriteMsg(&mockServer.WireReq); err != nil {
			t.Fatalf("WriteMsg failed[%s]", err)
		}
		if err := mockServer.Encoder.Flush(); err != nil {
			t.Fatalf("Write flush failed: [%s]", err)
		}
		log.Print("write 2 response")

		// 3 round - fetch
		if err := mockServer.Decoder.ReadMsg(&mockServer.WireRes); err != nil {
			t.Fatalf("ReadMsg failed[%s]", err)
		}

		method, ok = mockServer.WireRes.Response.(*dbp.Method)
		if !ok {
			t.Fatalf("received non-method message type:[%s]", mockServer.WireRes.MsgType)
		}

		blockArg = &lws.GetBlocksArg{}
		err = ptypes.UnmarshalAny(method.Params, blockArg)
		if err != nil {
			t.Fatalf("received non-getblocks message")
		}

		hash = []byte("1000000000000000000000000004")
		if bytes.Compare(blockArg.Hash, hash) != 0 {
			t.Fatalf("3 round fetch hash expect [%s], but [%s]", hash, blockArg.Hash)
		}

		if blockArg.Number != 1 {
			t.Fatalf("3 round fetch number expect 1, but [%d]", blockArg.Number)
		}

		results = []*lws.Block{
			&lws.Block{
				NHeight:  3,
				Hash:     []byte("1000000000000000000000000004"),
				HashPrev: []byte("0000000000000000000000000003"),
			},
		}

		serializedResults = make([]*any.Any, len(results))
		for idx, block := range results {
			anyAny, err := ptypes.MarshalAny(block)
			if err != nil {
				log.Fatal("could not serialize any field")
			}

			serializedResults[idx] = anyAny
		}

		resultMsg = &dbp.Result{
			Error:  "",
			Result: serializedResults,
		}

		mockServer.WireReq.ID = method.Id
		mockServer.WireReq.Request = resultMsg
		if err := mockServer.Encoder.WriteMsg(&mockServer.WireReq); err != nil {
			t.Fatalf("WriteMsg failed[%s]", err)
		}
		if err := mockServer.Encoder.Flush(); err != nil {
			t.Fatalf("Write flush failed: [%s]", err)
		}

		// 4 round - fetch
		if err := mockServer.Decoder.ReadMsg(&mockServer.WireRes); err != nil {
			t.Fatalf("ReadMsg failed[%s]", err)
		}

		method, ok = mockServer.WireRes.Response.(*dbp.Method)
		if !ok {
			t.Fatalf("received non-method message type:[%s]", mockServer.WireRes.MsgType)
		}

		blockArg = &lws.GetBlocksArg{}
		err = ptypes.UnmarshalAny(method.Params, blockArg)
		if err != nil {
			t.Fatalf("received non-getblocks message")
		}

		hash = []byte("0000000000000000000000000003")
		if bytes.Compare(blockArg.Hash, hash) != 0 {
			t.Fatalf("3 round fetch hash expect [%s], but [%s]", hash, blockArg.Hash)
		}

		if blockArg.Number != 1 {
			t.Fatalf("3 round fetch number expect 1, but [%d]", blockArg.Number)
		}

		results = []*lws.Block{
			&lws.Block{
				NHeight:  2,
				Hash:     []byte("0000000000000000000000000003"),
				HashPrev: []byte("0000000000000000000000000002"),
			},
		}

		serializedResults = make([]*any.Any, len(results))
		for idx, block := range results {
			anyAny, err := ptypes.MarshalAny(block)
			if err != nil {
				log.Fatal("could not serialize any field")
			}

			serializedResults[idx] = anyAny
		}

		resultMsg = &dbp.Result{
			Error:  "",
			Result: serializedResults,
		}

		mockServer.WireReq.ID = method.Id
		mockServer.WireReq.Request = resultMsg
		if err := mockServer.Encoder.WriteMsg(&mockServer.WireReq); err != nil {
			t.Fatalf("WriteMsg failed[%s]", err)
		}
		if err := mockServer.Encoder.Flush(); err != nil {
			t.Fatalf("Write flush failed: [%s]", err)
		}

		// 5 round - fetch
		if err := mockServer.Decoder.ReadMsg(&mockServer.WireRes); err != nil {
			t.Fatalf("ReadMsg failed[%s]", err)
		}

		method, ok = mockServer.WireRes.Response.(*dbp.Method)
		if !ok {
			t.Fatalf("received non-method message type:[%s]", mockServer.WireRes.MsgType)
		}

		blockArg = &lws.GetBlocksArg{}
		err = ptypes.UnmarshalAny(method.Params, blockArg)
		if err != nil {
			t.Fatalf("received non-getblocks message")
		}

		log.Print("final block args")

		hash = []byte("0000000000000000000000000002")
		if bytes.Compare(blockArg.Hash, hash) != 0 {
			log.Print("final hash error")
			t.Fatalf("5 round fetch hash expect [%s], but [%s]", hash, blockArg.Hash)
		}

		if blockArg.Number == 1 {
			log.Print("final number error")
			t.Fatalf("5 round fetch number not expect 1")
		}

		results = []*lws.Block{
			&lws.Block{
				NHeight:  1,
				Hash:     []byte("0000000000000000000000000002"),
				HashPrev: []byte("0000000000000000000000000001"),
			},
			&lws.Block{
				NHeight:  2,
				Hash:     []byte("0000000000000000000000000003"),
				HashPrev: []byte("0000000000000000000000000002"),
			},
			&lws.Block{
				NHeight:  3,
				Hash:     []byte("1000000000000000000000000004"),
				HashPrev: []byte("0000000000000000000000000003"),
			},
			&lws.Block{
				NHeight:  4,
				Hash:     []byte("1000000000000000000000000005"),
				HashPrev: []byte("1000000000000000000000000004"),
			},
			&lws.Block{
				NHeight:  5,
				Hash:     []byte("1000000000000000000000000006"),
				HashPrev: []byte("1000000000000000000000000005"),
			},
			&lws.Block{
				NHeight:  6,
				Hash:     []byte("1000000000000000000000000007"),
				HashPrev: []byte("1000000000000000000000000006"),
			},
		}

		serializedResults = make([]*any.Any, len(results))
		for idx, block := range results {
			anyAny, err := ptypes.MarshalAny(block)
			if err != nil {
				log.Fatal("could not serialize any field")
			}

			serializedResults[idx] = anyAny
		}

		resultMsg = &dbp.Result{
			Error:  "",
			Result: serializedResults,
		}

		mockServer.WireReq.ID = method.Id
		mockServer.WireReq.Request = resultMsg
		if err := mockServer.Encoder.WriteMsg(&mockServer.WireReq); err != nil {
			t.Fatalf("WriteMsg failed[%s]", err)
		}
		if err := mockServer.Encoder.Flush(); err != nil {
			t.Fatalf("Write flush failed: [%s]", err)
		}

	})(serverConn)

	client := &coreclient.Client{
		Addr: "whatever",
		Dial: func(addr string) (conn io.ReadWriteCloser, err error) {
			return clientConn, nil
		},
		DisableNegotiation: true,
	}
	cclientModule.SetPrimaryClient(client)
	client.Start()

	triggerBlock := &lws.Block{
		NHeight:  6,
		Hash:     []byte("1000000000000000000000000007"),
		HashPrev: []byte("1000000000000000000000000006"),
	}

	handleSyncBlock(triggerBlock, true)

	tail := GetTailBlock()
	tailHash := []byte("1000000000000000000000000007")
	if bytes.Compare(tail.Hash, tailHash) != 0 {
		t.Fatalf("tail hash expect [%s], but [%s]", tailHash, tail.Hash)
	}
}