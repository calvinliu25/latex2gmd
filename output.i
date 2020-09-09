%module output
%{
#include "output.h"
%}

%include <std_vector.i>
%include <std_string.i>
%template(stringVector) std::vector<std::string>;
%template(boolVector) std::vector<bool>;

%include "output.h"