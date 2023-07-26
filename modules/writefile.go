package modules

import (
    "log"
    "os"
)

func CreateNewDirectory() {
        if _, err := os.Stat("images"); err != nil {
        if os.IsNotExist(err) {
            if err := os.Mkdir("images", os.ModePerm);
            err != nil {
                log.Fatal(err)
            }
        }
    }
}
