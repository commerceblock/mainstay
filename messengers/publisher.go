package messengers

// Publisher interface
type Publisher interface {
    SendMessage(string, string)
    Close()
}
