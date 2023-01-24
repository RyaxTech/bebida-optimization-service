basename = slides

pdf: $(basename).pdf

$(basename).pdf: $(pdf) $(basename).tex
	rubber --pdf $(basename).tex

clean:
	rubber --pdf --clean $(basename).tex
	rm bibliography.bib
	rm $(basename).tex

mrproper:
	rm -f $(basename).pdf

disclean: mrproper
