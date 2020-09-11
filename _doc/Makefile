out_dir := build
input := frontmatter.md intro.md wxpay.md alipay.md paywall.md internal.md
output := subscripiton_api

.PHONY: build createdir clean
build : createdir
	pandoc -s --toc --pdf-engine=xelatex -o $(out_dir)/$(output).pdf $(input)

createdir :
	mkdir -p $(out_dir)

iap: createdir
	pandoc -s --toc --pdf-engine=xelatex -o $(out_dir)/apple_iap.pdf frontmatter.md iap.md

clean :
	rm build/*
