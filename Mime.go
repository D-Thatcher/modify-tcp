package main

func mime() {
	// buf := new(bytes.Buffer)
	// var r, reader io.Reader

	// //var chunked = false

	// if mimeHeader.Get("Transfer-Encoding") == "chunked" {
	// 	r = internal.NewChunkedReader(b)
	// } else {
	// 	r = b
	// }

	// switch mimeHeader.Get("Content-Encoding") {
	// case "gzip":
	// 	reader, err = gzip.NewReader(r)
	// 	if err != nil {
	// 		reader = r
	// 	}
	// default:
	// 	reader = r
	// }
	// buf.ReadFrom(reader)
	// body = buf.String()

}

func mimeother() {
	// var r io.Reader

	// conf := &config{}
	// for _, o := range opts {
	// 	o(conf)
	// }

	// br := bytes.NewReader(mv.message)
	// r = io.NewSectionReader(br, mv.bodyoffset, mv.traileroffset-mv.bodyoffset)

	// if !conf.decode {
	// 	return ioutil.NopCloser(r), nil
	// }

	// if mv.chunked {
	// 	r = httputil.NewChunkedReader(r)
	// }
	// switch mv.compress {
	// case "gzip":
	// 	gr, err := gzip.NewReader(r)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	return gr, nil
	// case "deflate":
	// 	return flate.NewReader(r), nil
	// default:
	// 	return ioutil.NopCloser(r), nil
	// }
}

// func main() {
// 	// a cached compressed content
// 	buf := bytes.NewBuffer(nil)

// 	// Create a flate.Writer
// 	flateWrite, err := flate.NewWriter(buf, flate.BestCompression)
// 	if err != nil {
// 		log.Fatalln(err)
// 	}
// 	defer flateWrite.Close()
// 	// Write the content to be compressed
// 	flateWrite.Write([]byte("compress/flate\n"))
// 	flateWrite.Flush()
// 	fmt.Printf("compressed content: %s\n", buf)

// 	// Extract the content that has just been compressed
// 	flateReader := flate.NewReader(buf)
// 	defer flateWrite.Close()
// 	// output
// 	fmt.Print("Unpacked content:")
// 	io.Copy(os.Stdout, flateReader)
// }
