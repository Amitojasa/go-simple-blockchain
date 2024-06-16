package main

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type Book struct{
	ID				string		`json:"id"`
	Title			string		`json:"title"`
	Author			string		`json:"author"`
	PublishDate		string		`json:"publish_date"`
	ISBN			string		`json:"isbn"`

}

type BookCheckout struct{
	// ID				string		`json:"id"`
	BookID			string		`json:"book_id"`
	User			string		`json:"user"`
	CheckoutDate	string		`json:"checkout_date"`
	IsGenesis		bool		`json:"is_genesis"`
}

type Block struct{
	Pos			int
	Data		BookCheckout
	Timestamp	string
	PrevHash	string
	Hash		string
}

type Blockchain struct{
	blocks []*Block
}

func (b *Block)generateHash(){
	bytes,_:=json.Marshal(b.Data)

	data:=string(b.Pos)+b.Timestamp+string(bytes)+b.PrevHash

	hash:=sha256.New()

	hash.Write(([]byte(data)))
	b.Hash= hex.EncodeToString(hash.Sum(nil))
}

func createBlock(prevBlock *Block, checkoutItem BookCheckout)(*Block){
	block :=&Block{}
	block.Pos=prevBlock.Pos+1
	block.PrevHash=prevBlock.Hash
	block.Timestamp=time.Now().String()
	block.Data = checkoutItem

	block.generateHash()

	return  block
}

func (b * Block)validateHash(hash string) bool{
	// bytes,_:= json.Marshal(b.Data)
	// data := string(b.Pos)+ b.Timestamp + string(bytes)+ b.PrevHash

	// h:=sha256.New()
	// h.Write([]byte(data))

	b.generateHash()

	return b.Hash==hash


}


func validBlock(block, prevBlock *Block) bool {
	if prevBlock.Hash != block.PrevHash{
		return false
	}

	if !block.validateHash(block.Hash){
		return false
	}

	if block.Pos!= (prevBlock.Pos+1){
		return false
	}

	return true


}



var blockchain * Blockchain

func newBook(w http.ResponseWriter,r *http.Request){
	var book Book

	if err := json.NewDecoder(r.Body).Decode(&book); err!=nil{
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Could not create:%v",err)
		w.Write([]byte("Could not create new book"))
		return
	}

	h:=md5.New()
	io.WriteString(h,book.ISBN+book.PublishDate)
	book.ID=fmt.Sprintf("%x",h.Sum(nil))
	resp,err:=json.MarshalIndent(book,""," ")
	
	if err!=nil{
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Could not marshal payload: %v",err)
		w.Write([]byte("Could not save book Data"))
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)

}

func (bc * Blockchain)AddBlock(data BookCheckout){
	prevBlock:=bc.blocks[len(bc.blocks)-1]
	block:=createBlock(prevBlock, data)

	if validBlock(block,prevBlock){
		bc.blocks=append(bc.blocks, block)
	}
}

func writeBlock(w http.ResponseWriter,r * http.Request){

	var checkoutItem BookCheckout

	if err:=json.NewDecoder(r.Body).Decode(&checkoutItem); err!=nil{
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Could not wite the block %v",err)
		w.Write([]byte("Could not write a block"))
	}

	blockchain.AddBlock(checkoutItem)

}

func getBlockchain(w http.ResponseWriter,r * http.Request){

	jbytes,err:=json.MarshalIndent(blockchain.blocks,""," ")

	if err!=nil{
		w.WriteHeader(http.StatusInternalServerError);
		json.NewEncoder(w).Encode(err)
		return
	}

	io.WriteString(w,string(jbytes))

}

func GenesisBlock()*Block{
	return createBlock(&Block{},BookCheckout{IsGenesis: true})
}

func NewBlockchain() *Blockchain{
	return &Blockchain{[]*Block{GenesisBlock()}}
}

func main(){

	blockchain = NewBlockchain()

	r:=mux.NewRouter()
	r.HandleFunc("/",getBlockchain).Methods("GET")
	r.HandleFunc("/",writeBlock).Methods("POST")
	r.HandleFunc("/new",newBook).Methods("POST")

	go func(){
		for _,block :=range blockchain.blocks{
			fmt.Printf("prevHash: %x\n", block.PrevHash)
			bytes,_:=json.MarshalIndent(block.Data,""," ")
			fmt.Printf("Data: %v\n", string(bytes))
			fmt.Printf("Hash: %x\n",block.Hash)
			fmt.Println()

		}
	}()

	log.Println("Listening on Port 3000")

	log.Fatal(http.ListenAndServe(":3000",r))
}