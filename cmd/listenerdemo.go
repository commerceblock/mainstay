// Demo of the communication between listener and clients
package main

import (
    "fmt"
    "time"

    "mainstay/config"
    "mainstay/messengers"
)

const CLIENT_PUB = "123456789ABCDEF"

func client_task() {
    client := messengers.NewDealerZmq(config.MAIN_LISTENER_PORT, CLIENT_PUB, nil)
    defer client.Close()

    for {
        // Inform listener and send block hash
        client.SendMessage([]byte("357abd41543a09f9290ff4b4ae008e317f252b80c96492bd9f346cced0943a7f"))

        // Wait for attestation confirmation
        client.ReadMessage()    //  Envelope delimiter
        workload := client.ReadMessage()    // Response from listener
        if string(workload) == "CONFIRMED" {
            fmt.Printf("Received confirmation\n")
        }

        //  Sleep
        time.Sleep(10 * time.Second)
    }
}

func main() {
    listener := messengers.NewRouterZmq(config.MAIN_LISTENER_PORT, nil)
    defer listener.Close()

    // spawn client task sending hash to listener router
    go client_task()

    for {
        var hash []byte
        identity := listener.ReadMessage()  // get last client
        fmt.Printf("[Listener] Received message from %s\n", identity)
        if string(identity) == CLIENT_PUB { // verify it is the client we expect
            listener.ReadMessage()  //  Envelope delimiter
            hash = listener.ReadMessage()   //  Response from client
            fmt.Printf("[Listener] Message: %s\n", hash)

            // wait for attestation confirmation
            // send confirmation
            time.Sleep(5 * time.Second)
            listener.SendMessage(identity, []byte("CONFIRMED"))
        }
    }
}
