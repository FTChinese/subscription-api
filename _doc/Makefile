build_dir := build
input := frontmatter.md intro.md wxpay.md alipay.md paywall.md internal.md
output := subscripiton_api

.PHONY: build mkbuild clean
build : mkbuild
	pandoc -s --toc --pdf-engine=xelatex -o $(build_dir)/$(output).pdf $(input)

mkbuild :
	mkdir -p $(build_dir)

clean :
	rm build/*