package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/parallaxsecond/parsec-client-go/parsec"
	"github.com/parallaxsecond/parsec-client-go/parsec/algorithm"
	"go.uber.org/zap"
)

var clients map[string]*parsec.BasicClient

type paramAll struct {
	Name    string
	KeyName string
	Message string
	Sign    string
}

type rtnCode struct {
	Code int32
}

type rtnKey struct {
	Name       string
	ProviderID uint32
}

func responseError(c *gin.Context, code int32) {
	c.AbortWithStatusJSON(http.StatusInternalServerError, &rtnCode{
		Code: code,
	})
}

func checkParam(c *gin.Context, checkLevel int32) (*paramAll, bool) {
	// get request param
	var data, _ = c.GetRawData()
	var param paramAll
	var err = json.Unmarshal(data, &param)
	if err != nil {
		return nil, false
	}

	// check param is valid
	if len(param.Name) == 0 {
		return nil, false
	}
	if checkLevel >= 1 {
		if len(param.KeyName) == 0 {
			return nil, false
		}
	}
	if checkLevel >= 2 {
		if len(param.Message) == 0 {
			return nil, false
		}
	}
	if checkLevel >= 3 {
		if len(param.Sign) == 0 {
			return nil, false
		}
	}

	return &param, true
}

func getSignAttr() *parsec.KeyAttributes {
	return parsec.DefaultKeyAttribute().SigningKey()
}

func getEncryptAttr(isPair bool) *parsec.KeyAttributes {
	var isDecrypt bool
	var keyType *parsec.KeyType
	if isPair {
		isDecrypt = true
		keyType = parsec.NewKeyType().RsaKeyPair()
	} else {
		isDecrypt = false
		keyType = parsec.NewKeyType().RsaPublicKey()
	}

	return &parsec.KeyAttributes{
		KeyBits: 2048,
		KeyType: keyType,
		KeyPolicy: &parsec.KeyPolicy{
			KeyAlgorithm: algorithm.NewAsymmetricEncryption().RsaPkcs1V15Crypt(),
			KeyUsageFlags: &parsec.UsageFlags{
				Cache:         false,
				Copy:          false,
				Decrypt:       isDecrypt,
				Derive:        false,
				Encrypt:       true,
				Export:        false,
				SignHash:      false,
				SignMessage:   false,
				VerifyHash:    false,
				VerifyMessage: false,
			},
		},
	}
}

func InitParsec() {
	clients = make(map[string]*parsec.BasicClient)
}

// curl -v -d '{"Name": "GoClient"}' 127.0.0.1:8300/client
func ApiNewClient(c *gin.Context) {
	param, ok := checkParam(c, 0)
	if !ok {
		responseError(c, CODE_INVALID_PARAM)
		return
	}

	// if not found in cache
	if _, ok := clients[param.Name]; !ok {
		// new parsec, ProviderMBed, DirectAuthenticator
		cfg := parsec.NewClientConfig().
			Provider(parsec.ProviderMBed).
			Authenticator(parsec.NewDirectAuthenticator(param.Name))
		client, err := parsec.CreateConfiguredClient(cfg)
		if err != nil {
			zap.L().Error(err.Error())
			responseError(c, CODE_PARSEC_ERROR)
			return
		}
		// ping to check if ok
		majver, minver, err := client.Ping()
		if err != nil {
			zap.L().Error(err.Error())
			client.Close()
			responseError(c, CODE_PARSEC_ERROR)
			return
		}
		if majver != 1 && minver != 0 {
			str := fmt.Sprintf("Parsec server version %v,%v was not supported!", majver, minver)
			zap.L().Error(str)
			client.Close()
			responseError(c, CODE_PARSEC_ERROR)
			return
		}
		// cache it
		clients[param.Name] = client
	}

	c.Status(http.StatusOK)
}

// curl -v -X DELETE -d '{"Name": "GoClient"}' 127.0.0.1:8300/client
func ApiDeleteClient(c *gin.Context) {
	param, ok := checkParam(c, 0)
	if !ok {
		responseError(c, CODE_INVALID_PARAM)
		return
	}

	// if found in cache
	if client, ok := clients[param.Name]; ok {
		client.Close()
		delete(clients, param.Name)
	}

	c.Status(http.StatusOK)
}

// curl -v -X GET -d '{"Name": "GoClient"}' 127.0.0.1:8300/keys
func ApiGetKeys(c *gin.Context) {
	param, ok := checkParam(c, 0)
	if !ok {
		responseError(c, CODE_INVALID_PARAM)
		return
	}

	// get client
	client, ok := clients[param.Name]
	if !ok {
		responseError(c, CODE_INVALID_CLIENT)
		return
	}

	keys, err := client.ListKeys()
	if err != nil {
		zap.L().Error(err.Error())
		responseError(c, CODE_PARSEC_ERROR)
		return
	}

	sliceKeys := make([]rtnKey, 0)
	for _, key := range keys {
		sliceKeys = append(sliceKeys, rtnKey{
			Name:       key.Name,
			ProviderID: uint32(key.ProviderID),
		})
	}

	c.JSON(http.StatusOK, sliceKeys)
}

// curl -v -X DELETE -d '{"Name": "GoClient"}' 127.0.0.1:8300/keys
func ApiDeleteKeys(c *gin.Context) {
	param, ok := checkParam(c, 0)
	if !ok {
		responseError(c, CODE_INVALID_PARAM)
		return
	}

	// get client
	client, ok := clients[param.Name]
	if !ok {
		responseError(c, CODE_INVALID_CLIENT)
		return
	}

	keys, err := client.ListKeys()
	if err != nil {
		zap.L().Error(err.Error())
		responseError(c, CODE_PARSEC_ERROR)
		return
	}

	for _, key := range keys {
		client.PsaDestroyKey(key.Name)
	}

	c.Status(http.StatusOK)
}

func newKey(c *gin.Context, isSign bool) {
	param, ok := checkParam(c, 1)
	if !ok {
		responseError(c, CODE_INVALID_PARAM)
		return
	}

	// get client
	client, ok := clients[param.Name]
	if !ok {
		responseError(c, CODE_INVALID_CLIENT)
		return
	}

	var keyAttr *parsec.KeyAttributes
	if isSign {
		keyAttr = getSignAttr()
	} else {
		keyAttr = getEncryptAttr(true)
	}

	// new key
	err := client.PsaGenerateKey(param.KeyName, keyAttr)
	if err != nil {
		zap.L().Error(err.Error())
		responseError(c, CODE_PARSEC_ERROR)
		return
	}

	c.Status(http.StatusOK)
}

// curl -v -d '{"Name": "GoClient", "KeyName": "MyKey"}' 127.0.0.1:8300/keysign
func ApiNewSignKey(c *gin.Context) {
	newKey(c, true)
}

// curl -v -d '{"Name": "GoClient", "KeyName": "MyEncKey"}' 127.0.0.1:8300/keyenc
func ApiNewEncKey(c *gin.Context) {
	newKey(c, false)
}

// curl -v -d '{"Name": "GoClient", "KeyName": "MyPubKey", "Message":"ssh-rsa xxx keyname"}' 127.0.0.1:8300/key
func ApiSetKeyPub(c *gin.Context) {
	param, ok := checkParam(c, 2)
	if !ok {
		responseError(c, CODE_INVALID_PARAM)
		return
	}

	// get client
	client, ok := clients[param.Name]
	if !ok {
		responseError(c, CODE_INVALID_CLIENT)
		return
	}

	// ssh-rsa pub -> []byte
	strs := strings.Split(param.Message, " ")
	if len(strs) != 3 {
		responseError(c, CODE_INVALID_PARAM)
		return
	}

	if strings.TrimSpace(strs[0]) != "ssh-rsa" {
		responseError(c, CODE_INVALID_PARAM)
		return
	}

	pubkey, err := base64.StdEncoding.DecodeString(strings.TrimSpace(strs[1]))
	if err != nil {
		zap.L().Error(err.Error())
		responseError(c, CODE_INVALID_PARAM)
		return
	}

	keyAttr := getEncryptAttr(false)
	err = client.PsaImportKey(param.KeyName, keyAttr, pubkey)
	if err != nil {
		zap.L().Error(err.Error())
		responseError(c, CODE_PARSEC_ERROR)
		return
	}

	c.Status(http.StatusOK)
}

// curl -v -X GET -d '{"Name": "GoClient", "KeyName": "MyEncKey"}' 127.0.0.1:8300/key
func ApiGetKeyPub(c *gin.Context) {
	param, ok := checkParam(c, 1)
	if !ok {
		responseError(c, CODE_INVALID_PARAM)
		return
	}

	// get client
	client, ok := clients[param.Name]
	if !ok {
		responseError(c, CODE_INVALID_CLIENT)
		return
	}

	// get key
	b, err := client.PsaExportPublicKey(param.KeyName)
	if err != nil {
		zap.L().Error(err.Error())
		responseError(c, CODE_PARSEC_ERROR)
		return
	}

	// []byte -> ssh-rsa pub
	str := base64.StdEncoding.EncodeToString(b)
	str = "ssh-rsa " + str + " " + param.Name + "_" + param.KeyName
	c.String(http.StatusOK, str)
}

// curl -v -X DELETE -d '{"Name": "GoClient", "KeyName": "MyKey"}' 127.0.0.1:8300/key
func ApiDeleteKey(c *gin.Context) {
	param, ok := checkParam(c, 1)
	if !ok {
		responseError(c, CODE_INVALID_PARAM)
		return
	}

	// get client
	client, ok := clients[param.Name]
	if !ok {
		responseError(c, CODE_INVALID_CLIENT)
		return
	}

	client.PsaDestroyKey(param.KeyName)

	c.Status(http.StatusOK)
}

// curl -v -d '{"Name": "GoClient", "KeyName": "MyKey", "Message": "Hello World"}' 127.0.0.1:8300/sign
func ApiSign(c *gin.Context) {
	param, ok := checkParam(c, 2)
	if !ok {
		responseError(c, CODE_INVALID_PARAM)
		return
	}

	// get client
	client, ok := clients[param.Name]
	if !ok {
		responseError(c, CODE_INVALID_CLIENT)
		return
	}

	// ONLY support sign hash
	hash, err := client.PsaHashCompute([]byte(param.Message), algorithm.HashAlgorithmTypeSHA256)
	if err != nil {
		responseError(c, CODE_PARSEC_ERROR)
		return
	}

	// sign
	keyAttr := getSignAttr()
	keyalg := keyAttr.KeyPolicy.KeyAlgorithm.GetAsymmetricSignature()
	signature, err := client.PsaSignHash(param.KeyName, hash, keyalg)
	if err != nil {
		zap.L().Error(err.Error())
		responseError(c, CODE_PARSEC_ERROR)
		return
	}

	// []byte -> base64
	str := base64.StdEncoding.EncodeToString(signature)
	c.String(http.StatusOK, str)
}

// curl -v -d '{"Name": "GoClient", "KeyName": "MyKey", "Message": "Hello World", "Sign": "xxx"}' 127.0.0.1:8300/verify
func ApiVerify(c *gin.Context) {
	param, ok := checkParam(c, 3)
	if !ok {
		responseError(c, CODE_INVALID_PARAM)
		return
	}

	// get client
	client, ok := clients[param.Name]
	if !ok {
		responseError(c, CODE_INVALID_CLIENT)
		return
	}

	hash, err := client.PsaHashCompute([]byte(param.Message), algorithm.HashAlgorithmTypeSHA256)
	if err != nil {
		zap.L().Error(err.Error())
		responseError(c, CODE_PARSEC_ERROR)
		return
	}

	signature, err := base64.StdEncoding.DecodeString(param.Sign)
	if err != nil {
		zap.L().Error(err.Error())
		responseError(c, CODE_INVALID_PARAM)
		return
	}

	keyAttr := getSignAttr()
	keyalg := keyAttr.KeyPolicy.KeyAlgorithm.GetAsymmetricSignature()

	err = client.PsaVerifyHash(param.KeyName, hash, signature, keyalg)
	if err != nil {
		zap.L().Error(err.Error())
		responseError(c, CODE_VERIFY_FAIL)
		return
	}

	c.Status(http.StatusOK)
}

// curl -v -d '{"Name": "GoClient", "KeyName": "MyEncKey", "Message": "Hello World"}' 127.0.0.1:8300/encrypt
func ApiEncrypt(c *gin.Context) {
	param, ok := checkParam(c, 2)
	if !ok {
		responseError(c, CODE_INVALID_PARAM)
		return
	}

	// get client
	client, ok := clients[param.Name]
	if !ok {
		responseError(c, CODE_INVALID_CLIENT)
		return
	}

	keyAttr := getEncryptAttr(true)
	keyalg := keyAttr.KeyPolicy.KeyAlgorithm.GetAsymmetricEncryption()

	ciphertext, err := client.PsaAsymmetricEncrypt(param.KeyName, keyalg, []byte{}, []byte(param.Message))
	if err != nil {
		zap.L().Error(err.Error())
		responseError(c, CODE_PARSEC_ERROR)
		return
	}

	// []byte -> base64
	str := base64.StdEncoding.EncodeToString(ciphertext)
	c.String(http.StatusOK, str)
}

// curl -v -d '{"Name": "GoClient", "KeyName": "MyEncKey", "Message": "xxxx"}' 127.0.0.1:8300/decrypt
func ApiDecrypt(c *gin.Context) {
	param, ok := checkParam(c, 2)
	if !ok {
		responseError(c, CODE_INVALID_PARAM)
		return
	}

	// get client
	client, ok := clients[param.Name]
	if !ok {
		responseError(c, CODE_INVALID_CLIENT)
		return
	}

	ciphertext, err := base64.StdEncoding.DecodeString(param.Message)
	if err != nil {
		zap.L().Error(err.Error())
		responseError(c, CODE_INVALID_PARAM)
		return
	}

	keyAttr := getEncryptAttr(true)
	keyalg := keyAttr.KeyPolicy.KeyAlgorithm.GetAsymmetricEncryption()

	plaintext, err := client.PsaAsymmetricDecrypt(param.KeyName, keyalg, []byte{}, ciphertext)
	if err != nil {
		zap.L().Error(err.Error())
		responseError(c, CODE_PARSEC_ERROR)
		return
	}

	c.String(http.StatusOK, string(plaintext))
}
