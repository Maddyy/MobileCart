package main

import "fmt"
import "encoding/json"
import "crypto/sha256"
import "encoding/hex"
import "time"
import "net/http"
import "io"
import "crypto/md5"
import "log"
import "github.com/gorilla/mux"

type Block struct {
	Pos int
	Data MobileCheckOut
	TimeStamp string
	Hash      string
	PrevHash  string
}

type MobileCheckOut struct {
MobileSerial string `json:"mobile_serial"`
UserName   string `json:"user_name"`
BuyDate    string `json:"checkout_date"`
CheckGenesis bool `json:"check_genesis"`
}

type Mobile struct {
	IMEI          string `json:"imei"`
	Company       string `json:"company"`
	Model      string `json:"model"`
	ManufacturerDate string `json:"publish_date"`
	Country        string `json:"country"`
}

func (b *Block) generateHash(){
	bytes,_ :=json.Marshal(b.Data)
	data := string(b.Pos)+b.TimeStamp+string(bytes)+b.PrevHash
	hash := sha256.New()
	hash.Write([]byte(data))
	b.Hash = hex.EncodeToString(hash.Sum(nil))
}
func CreateBlock(prevBlock *Block,buyMobileItem MobileCheckOut) *Block{
	block := &Block{}
	block.Pos = prevBlock.Pos+1
	block.TimeStamp = time.Now().String()
	block.Data = buyMobileItem
	block.PrevHash = prevBlock.Hash
	block.generateHash()
	return block
}
type Blockchain struct {
	blocks []*Block
}
var BlockChain *Blockchain

func (bc *Blockchain) AddBlock(data MobileCheckOut){
	prevBlock :=bc.blocks[len(bc.blocks)-1]
	block := CreateBlock(prevBlock,data)
	if validBlock(block, prevBlock) {
		bc.blocks = append(bc.blocks, block)
	  }
}
func GenesisBlock() *Block {
	return CreateBlock(&Block{}, MobileCheckOut{CheckGenesis: true})
  }
  func NewBlockchain() *Blockchain {
	return &Blockchain{[]*Block{GenesisBlock()}}
  }
  func validBlock(block, prevBlock *Block) bool {
	if prevBlock.Hash != block.PrevHash {
	  return false
	}
	if !block.validateHash(block.Hash) {
	  return false
	}
	if prevBlock.Pos+1 != block.Pos {
	  return false
	}
	return true
  }
  func (b *Block) validateHash(hash string) bool {
	b.generateHash()
	if b.Hash != hash {
	  return false
	}
	return true
  }
func getBlockchain(w http.ResponseWriter, r *http.Request){
	 jasonbytes,err := json.MarshalIndent(BlockChain.blocks,"","")
	 if err != nil {
		 w.WriteHeader(http.StatusInternalServerError)
		 json.NewEncoder(w).Encode(err)
         return
	 } 
	 io.WriteString(w, string(jasonbytes))
}
func writeBlock(w http.ResponseWriter, r *http.Request){
  var checkoutMobile MobileCheckOut
  if err := json.NewDecoder(r.Body).Decode(&checkoutMobile); err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    log.Printf("could not write Block: %v", err)
    w.Write([]byte("could not write block"))
	return
  }
  BlockChain.AddBlock(checkoutMobile)
  resp, err := json.MarshalIndent(checkoutMobile, "", " ")
  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    log.Printf("could not marshal payload: %v", err)
    w.Write([]byte("could not write block"))
    return
  }
  w.WriteHeader(http.StatusOK)
  w.Write(resp)
}
func newMobile(w http.ResponseWriter, r *http.Request){
	var mobile Mobile
	if err := json.NewDecoder(r.Body).Decode(&mobile); err != nil {
	  w.WriteHeader(http.StatusInternalServerError)
	  log.Printf("could not create: %v", err)
	  w.Write([]byte("could not create new mobile"))
	  return
	}
	h := md5.New()
	io.WriteString(h, mobile.Model+mobile.ManufacturerDate)
    mobile.IMEI = fmt.Sprintf("%x", h.Sum(nil))
    resp, err := json.MarshalIndent(mobile, "", " ")
  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    log.Printf("could not marshal payload: %v", err)
    w.Write([]byte("could not save mobile data"))
	return
  }
  w.WriteHeader(http.StatusOK)
  w.Write(resp)
}
func main(){
	BlockChain = NewBlockchain()

	r := mux.NewRouter()
	r.HandleFunc("/", getBlockchain).Methods("GET")
	r.HandleFunc("/", writeBlock).Methods("POST")
	r.HandleFunc("/new", newMobile).Methods("POST")
	
	go func() {
		for _, block := range BlockChain.blocks {
		  fmt.Printf("Prev. hash: %x\n", block.PrevHash)
		  bytes, _ := json.MarshalIndent(block.Data, "", " ")
		  fmt.Printf("Data: %v\n", string(bytes))
		  fmt.Printf("Hash: %x\n", block.Hash)
		  fmt.Println()
		}
	  }()
	  log.Println("Listening on port 7001")
	
	  log.Fatal(http.ListenAndServe(":7001", r))
}