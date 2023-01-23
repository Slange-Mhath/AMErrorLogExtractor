<!-- ABOUT THE PROJECT -->
## About AM Errorlog Extractor     
[![Golang][Golang]][Golang-url]

This tool scrapes the error log of Archivematica and extracts error messages which contain user provided keywords. 

As we were using Archivematica mostly independent of the Dashboard we are lacking a way to pick up on potential errors 
during the preservation process. Extracting the errors from the DB, with Keywords we are able to provide and feeding 
the resulting JSON file into Grapha will overcome this problem and give us an idea where the process failed.

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- GETTING STARTED -->
## Getting Started



### Prerequisites

Getting Archivematica up and running is out of the scope of this document. Please refer to the [Archivematica GitHub]
for further information.

Check you have the latest version of Go.
* golang
  ```
  $ go version
  ```


### Installation


1. Clone the repo
   ```sh
   git clone https://github.com/Slange-Mhath/AMErrorLogExtractor.git
   ```

2. Navigate into the project directory
   ```sh
   cd AMErrorLogExtractor
   ```

3. Use the go build command to build the project
   ```sh
   go build
   ```
   
4. Use ./AMErrorLogExtractor to run the project
   ```sh
    ./AMErrorLogExtractor
   ```
    Alternativeliely you can use go run to run the project
   ```sh
   go run main.go
   ```


<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- USAGE EXAMPLES -->
## Usage

Provide the following parameters to the tool:

1. dbPassword: The password of the Archivematica database user
2. dbUser: The Archivematica database user
3. dbNet: The network address of the Archivematica database
4. ipAddr: The IP address of the Archivematica server
5. dbName: The name of the Archivematica database
6. keyFile: The path to the file containing the keywords
7. outputFile: The path to the output file

So an example would probably look like that: 

   ```sh
    go run . -dbPass=mypassword -dbUser=myuser -dbNet=tcp -ipAddr=127.0.0.1:62001 -dbName=mydb -keyFile=keywords.txt -outputFile=output.json
   ```



<!-- LICENSE -->
## License

Distributed under the MIT License. See `LICENSE.txt` for more information.

<p align="right">(<a href="#readme-top">back to top</a>)</p>
















[Golang]: https://img.shields.io/badge/Made%20with-Go-1f425f.svg
[Golang-url]: https://go.dev/
[Archivematica GitHub]: https://github.com/artefactual/archivematica

