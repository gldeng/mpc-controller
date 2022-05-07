package core

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	pkgErrors "github.com/pkg/errors"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

type KeygenRequest struct {
	RequestId       string   `json:"request_id"`
	ParticipantKeys []string `json:"public_keys"`
	Threshold       uint64   `json:"t"`
}

type SignRequest struct {
	RequestId       string   `json:"request_id"`
	PublicKey       string   `json:"public_key"`
	ParticipantKeys []string `json:"participant_public_keys"`
	Hash            string   `json:"message"`
}

type Result struct {
	RequestId     string `json:"request_id"`
	Result        string `json:"result"`
	RequestType   string `json:"request_type"`
	RequestStatus string `json:"request_status"`
}

var _ MPCClient = (*MpcClient)(nil)

type MPCClient interface {
	Keygen(ctx context.Context, keygenReq *KeygenRequest) error
	Sign(ctx context.Context, signReq *SignRequest) error
	Result(ctx context.Context, reqID string) (*Result, error)
}

type MpcClient struct {
	url string
}

func NewMpcClient(url string) (*MpcClient, error) {
	return &MpcClient{url: url}, nil
}

func (c *MpcClient) Keygen(ctx context.Context, request *KeygenRequest) error {
	normalized, err := normalizePubKeys(request.ParticipantKeys)
	if err != nil {
		return err
	}
	request.ParticipantKeys = normalized
	payloadBytes, err := json.Marshal(request)
	if err != nil {
		return err
	}

	res, err := http.Post(c.url+"/keygen", "application/json", bytes.NewBuffer(payloadBytes))
	fmt.Printf("response is %v\n", res)
	if err != nil {
		fmt.Printf("err is %v\n", err)
		return err
	}
	return nil

}

func (c *MpcClient) Sign(ctx context.Context, request *SignRequest) error {
	//normalized, err := normalizePubKeys(request.ParticipantKeys)
	//fmt.Printf("normalized keys %v\n", normalized)
	//if err != nil {
	//	log.Fatalf("%+v", pkgErrors.WithStack(err))
	//	return pkgErrors.WithStack(err)
	//}
	//request.ParticipantKeys = normalized
	payloadBytes, err := json.Marshal(request)
	if err != nil {
		return err
	}

	res, err := http.Post(c.url+"/sign", "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	fmt.Printf("response is %v\n", res)
	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	bodyString := string(bodyBytes)
	fmt.Printf("body is %v\n", bodyString)
	if err != nil {
		return err
	}
	return nil
}

func (c *MpcClient) Result(ctx context.Context, requestId string) (*Result, error) {
	payload := strings.NewReader("")
	res, err := http.Post(c.url+"/result/"+requestId, "application/json", payload)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	var result Result
	body, _ := ioutil.ReadAll(res.Body)
	err = json.Unmarshal(body, &result)
	return &result, nil
}

func normalizePubKey(pubKey string) (*string, error) {
	pub := common.Hex2Bytes(pubKey)
	pubKey0 := pubKey[0]

	if len(pub) == 33 && (pubKey0 == 3) || (pubKey0 == 2) {
		// Compressed format
		return &pubKey, nil
	} else if len(pub) == 65 && pubKey[0] == 4 {
		compressed, err := toCompressed(pub)
		if err != nil {
			return nil, err
		}
		pubN := common.Bytes2Hex(compressed)
		return &pubN, nil
	} else if len(pub) == 64 {
		var newPub [65]byte
		newPub[0] = 4
		copy(newPub[1:], pub)
		compressed, err := toCompressed(newPub[:])
		if err != nil {
			return nil, err
		}
		pubN := common.Bytes2Hex(compressed)
		return &pubN, nil
	} else {
		return nil, pkgErrors.New("invalid secp256k1 public key")
	}
}

func toCompressed(pub []byte) ([]byte, error) {
	x, y := elliptic.Unmarshal(crypto.S256(), pub)
	if x == nil {
		return nil, pkgErrors.New("invalid secp256k1 public key")
	}
	pk := &ecdsa.PublicKey{Curve: crypto.S256(), X: x, Y: y}
	pubCompressed := elliptic.MarshalCompressed(crypto.S256(), pk.X, pk.Y)
	return pubCompressed, nil
}

func normalizePubKeys(pubKeys []string) ([]string, error) {
	var out []string
	for _, hex := range pubKeys {
		normalized, err := normalizePubKey(hex)
		if err != nil {
			return nil, err
		}
		out = append(out, *normalized)
	}
	return out, nil
}
