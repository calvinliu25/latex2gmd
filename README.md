# Latex 2 Github Flavored Markdown

Latex2gmd is a program that will convert your Latex file (.tex) into Github Flavored Markdown files (.md).
Latex2gmd will be able to handle a subset of common Latex math display formats and render it into .md files.
sample.tex is provided as an example of what latex2gmd can process and acts as a test file for the program.

To run the program, use ./latex2gmd.sh \<filename\>.tex \<filename\>.md

To compile the SWIG interface, use "./compileswig.sh" or the following commands as an alternative:

swig -c++ -python -py3 -modern output.i

g++ -O2 -fPIC -c output.cpp output_wrap.cxx -I/usr/include/python3.6

g++ -shared output.o output_wrap.o -o _output.so

To cleanup the SWIG interface, use "./cleanup.sh"

Refer to help.txt for more details
