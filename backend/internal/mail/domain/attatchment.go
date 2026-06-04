package domain



type Attatchment struct{
	Filename string
    ContentType string            // "application/pdf", "application/zip", ...
    Bytes []uint8
}
