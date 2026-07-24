package idgen

import gonanoid "github.com/matoous/go-nanoid/v2"

const urlAlphabet = "useandom-26T198340PX75pxJACKVERYMINDBUSHWOLF_GQZbfghjklqvwyzrict"

func GenerateID() (string, error) {
	id, err := gonanoid.Generate(urlAlphabet, 16)
	if err != nil {
		return "", err
	}
	return id, nil
}
