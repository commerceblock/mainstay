package messengers

// Subscriber Interface
type Subscriber interface {
    ReadMessage()   (string, string)
    Close()
}
