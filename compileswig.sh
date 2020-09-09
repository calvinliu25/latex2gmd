#!/bin/bash
# To compile the SWIG interface use the following commands:

swig -c++ -python -py3 -modern output.i
g++ -O2 -fPIC -c output.cpp output_wrap.cxx -I/usr/include/python3.6
g++ -shared output.o output_wrap.o -o _output.so
