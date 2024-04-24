package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"

	"github.com/keep-starknet-strange/art-peace/backend/core"
)

func InitNFTRoutes() {
	http.HandleFunc("/get-nft", getNFT)
	http.HandleFunc("/get-nfts", getNFTs)
	http.HandleFunc("/get-my-nfts", getMyNFTs)
	http.HandleFunc("/mint-nft-devnet", mintNFTDevnet)
	// Create a static file server for the nft images
	http.Handle("/nft-images/", http.StripPrefix("/nft-images/", http.FileServer(http.Dir("."))))
	//http.HandleFunc("/nft-image", nftImage)
}

type NFTData struct {
	TokenID     int    `json:"tokenId"`
	Position    int    `json:"position"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	ImageHash   string `json:"imageHash"`
	BlockNumber int    `json:"blockNumber"`
	Minter      string `json:"minter"`
}

type NFTLikes struct {
	nftkey	int 			`json:"position"`
	liker 	string 		`json:"height"`
}


func getNFT(w http.ResponseWriter, r *http.Request) {
	tokenId := r.URL.Query().Get("tokenId")

	var nftData NFTData
	rows, err := core.ArtPeaceBackend.Databases.Postgres.Query(context.Background(), "SELECT * FROM nfts WHERE token_id = $1", tokenId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	defer rows.Close()

	err = rows.Scan(&nftData.TokenID, &nftData.Position, &nftData.Width, &nftData.Height, &nftData.ImageHash, &nftData.BlockNumber, &nftData.Minter)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	out, err := json.Marshal(nftData)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Write([]byte(out))
}

func getMyNFTs(w http.ResponseWriter, r *http.Request) {
	address := r.URL.Query().Get("address")

	var nftDatas []NFTData
	rows, err := core.ArtPeaceBackend.Databases.Postgres.Query(context.Background(), "SELECT * FROM nfts WHERE minter = $1", address)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	defer rows.Close()

	for rows.Next() {
		var nftData NFTData
		err = rows.Scan(&nftData.TokenID, &nftData.Position, &nftData.Width, &nftData.Height, &nftData.ImageHash, &nftData.BlockNumber, &nftData.Minter)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		nftDatas = append(nftDatas, nftData)
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	out, err := json.Marshal(nftDatas)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Write([]byte(out))
}

func getNFTs(w http.ResponseWriter, r *http.Request) {
	// TODO: Pagination & Likes
	var nftDatas []NFTData
	rows, err := core.ArtPeaceBackend.Databases.Postgres.Query(context.Background(), "SELECT * FROM nfts")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	defer rows.Close()

	for rows.Next() {
		var nftData NFTData
		err = rows.Scan(&nftData.TokenID, &nftData.Position, &nftData.Width, &nftData.Height, &nftData.ImageHash, &nftData.BlockNumber, &nftData.Minter)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		nftDatas = append(nftDatas, nftData)
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	out, err := json.Marshal(nftDatas)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Write([]byte(out))
}

// func nftImage(w http.ResponseWriter, r *http.Request) {
//   // Get the png image at location "nft-{tokenId}.png"
//   tokenId := r.URL.Query().Get("tokenId")
//   imageLocation := fmt.Sprintf("nft-%s.png", tokenId)
//
//   image, err := os.Open(imageLocation)
//   if err != nil {
//     w.WriteHeader(http.StatusInternalServerError)
//     w.Write([]byte(err.Error()))
//     return
//   }
//   defer image.Close()
//
//   w.Header().Set("Access-Control-Allow-Origin", "*")
//   w.Header().Set("Content-Type", "image/png")
//   w.WriteHeader(http.StatusOK)
//
//   io.Copy(w, image)
// }

func mintNFTDevnet(w http.ResponseWriter, r *http.Request) {
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	var jsonBody map[string]string
	err = json.Unmarshal(reqBody, &jsonBody)
	if err != nil {
		panic(err)
	}

	position, err := strconv.Atoi(jsonBody["position"])
	if err != nil {
		panic(err)
	}

	width, err := strconv.Atoi(jsonBody["width"])
	if err != nil {
		panic(err)
	}

	height, err := strconv.Atoi(jsonBody["height"])
	if err != nil {
		panic(err)
	}

	shellCmd := core.ArtPeaceBackend.BackendConfig.Scripts.MintNFTDevnet
	contract := os.Getenv("ART_PEACE_CONTRACT_ADDRESS")

	cmd := exec.Command(shellCmd, contract, "mint_nft", strconv.Itoa(position), strconv.Itoa(width), strconv.Itoa(height))
	_, err = cmd.Output()
	if err != nil {
		fmt.Println("Error executing shell command: ", err)
		panic(err)
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write([]byte("Minted NFT on devnet"))
}

func LikeNFT(w http.ResponseWriter, r *http.Request) {
	nftkey := r.URL.Query().Get("nft_key")
	userAddress := r.URL.Query().Get("user_address")

	// check if the user has like the nft 

	var nftlike NFTLikes
	rows, err := core.ArtPeaceBackend.Databases.Postgres.Query(context.Background(), "SELECT * FROM nftlikes WHERE nft_key = $1 AND user_address = $2", nftkey, userAddress)
	
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	defer rows.Close()

	err = rows.Scan(&nftlike.nftkey, &nftlike.liker)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	out, err := json.Marshal(nftlike)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	if out {
			w.Write([]byte(out))
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("NFT Like By User"))
			fmt.Println("NFT Like By User")
			return 
	}


	 //  new like for nft 
		_, err = core.ArtPeaceBackend.Databases.Postgres.Exec(context.Background(), "INSERT INTO nftlikes (nft_key) (user_address) VALUES ($1) ($2)", nftkey, userAddress)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("NFT Like By User"))
	fmt.Println("NFT Like By User")
}


func UnlikeNft(w http.ResponseWriter, r *http.Request) {
	nftkey := r.URL.Query().Get("nft_key")
	userAddress := r.URL.Query().Get("user_address")

	// check if the user has like the nft 

	var nftlike NFTLikes
	rows, err := core.ArtPeaceBackend.Databases.Postgres.Query(context.Background(), "SELECT * FROM nftlikes WHERE nft_key = $1 AND user_address = $2", nftkey, userAddress)
	
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	defer rows.Close()

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	out, err := json.Marshal(nftlike)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	if out {
		// delete the like 
		rows, err := core.ArtPeaceBackend.Databases.Postgres.Query(context.Background(), "DELETE FROM nftlikes WHERE nft_key = $1 AND user_address = $2", nftkey, userAddress)
	
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		defer rows.Close()

		w.Write([]byte(out))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("NFT Unlike By User"))
		fmt.Println("NFT Unlike By User")
		return 

	}

	w.Write([]byte(out))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Cannot Unlike an NFT You did not Like"))
	fmt.Println("Cannot Unlike an NFT You did not Like")
	return 

}