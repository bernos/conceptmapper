all: clean build-example

clean:
	rm -rf examples/mkdocs/docs

build-example:
	go run main.go generate-markdown-site -o ./examples/mkdocs/docs ./examples/mkdocs/concept-map.yaml

serve-example:
	cd examples/mkdocs && mkdocs serve
