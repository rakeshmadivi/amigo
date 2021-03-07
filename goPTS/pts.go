package main
import(
	"log"
	"net/http"
	"io"
	//"io/ioutil"
	"os"
	"regexp"
)

func main(){
	setupRoutes()
}

func setupRoutes(){
	http.HandleFunc("/upload", uploadFile)
	http.HandleFunc("/setStatus", setStatus)
	http.HandleFunc("/getStatus", getStatus)

	port := "8300"
	
	log.Print("Listening on:", port)

	http.ListenAndServe(":" + port,nil)
}

func uploadFile(w http.ResponseWriter, r *http.Request){
	reg, _ := regexp.Compile(":.*$")

	remoteIP := reg.ReplaceAllString(r.RemoteAddr,"")

	log.Print("Upload Request from: ", remoteIP)
	r.ParseMultipartForm( 32<<20 )

	file, handler, err := r.FormFile("file")

	if err != nil {
		log.Print("Failed to parse file.")
		return
	}

	defer file.Close()

	// This is path which we want to store the file
	directory := "./uploaded-files/"
	
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		os.Mkdir(directory, 0755) //os.ModeDir)
	}
	
	f, err := os.OpenFile( directory + handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)

	if err != nil {
		log.Print(err) //"Error in opening file while uploading")
		return
	}

	// Copy the file to the destination path
	io.Copy(f, file)

}

func setStatus(w http.ResponseWriter, r *http.Request){
	log.Print("setStatus")
	log.Print("SRC-IP:",r.Header.Get("X-Forwarded-for"))
	log.Print("SRC-IP:",r.RemoteAddr)
}

func getStatus(w http.ResponseWriter, r *http.Request){
	log.Print("getStatus")
}
