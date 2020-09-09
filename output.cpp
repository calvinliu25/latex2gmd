// Output functions for latex2gmd by Calvin Liu
// Handles all the output formatting to Github Flavored Markdown
// These functions are called in Python through the SWIG interface
// adapted from https://stackoverflow.com/questions/5251042/passing-python-array-to-c-function-with-swig
// adapted from https://stackoverflow.com/questions/47933692/trouble-mapping-to-vectorstdstring

#include "output.h"

using namespace std;

// replaceString - will replace every instance of an existing substring with a new substring
// referenced https://stackoverflow.com/questions/2896600/how-to-replace-all-occurrences-of-a-character-in-string
static inline void replaceString(string &str, const string &from, const string &to)
{
    size_t startPos = 0;
    while ((startPos = str.find(from, startPos)) != string::npos)
    {
        str.replace(startPos, from.length(), to);
        startPos += to.length(); // Handles case where 'to' is a substring of 'from'
    }
}

// parseTokens - parses the tokens passed from the client into a .md format
string parseTokens(vector<string> dataTokens, vector<bool> mathTokens)
{
    // newLineFlag and mathFlag are internal toggleable flag that help keep track of formatting
    string result = "";
    string dataString = "";
    bool newLineFlag = false;
    bool mathFlag = false;

    // dataTokens.size() is same as mathTokens.size()
    for (int i = 0; i < dataTokens.size(); i++)
    {
        dataString = dataTokens[i];

        // math mode will render Latex using latex.codecogs.com
        // adapted from https://stackoverflow.com/questions/35498525/latex-rendering-in-readme-md-on-github
        if (mathFlag)
        {
            // check for math mode before processing
            if (mathTokens[i] == true)
            {
                mathFlag = false;
                continue;
            }

            if (dataTokens[i] == "")
            {
                if (newLineFlag)
                {
                    result = result + "\n";
                    newLineFlag = false;
                }
            }
            else
            {
                // math mode parsing of data which uses latex.codecogs.com
                replaceString(dataString, " ", "%20");
                replaceString(dataString, "&", "%20");
                replaceString(dataString, "$", "%20");
                replaceString(dataString, "\\sfrac", "\\frac");
                dataString = "![equation](http://latex.codecogs.com/gif.latex?" + dataString + ")";
                result = result + dataString + "\n\n";
                newLineFlag = true;
            }
        }
        else
        {
            // check for math mode before processing
            if (mathTokens[i] == true)
            {
                mathFlag = true;
                continue;
            }

            if (dataString == "")
            {
                if (newLineFlag)
                {
                    result = result + "\n";
                    newLineFlag = false;
                }
            }
            else
            {
                // regular parsing of data
                replaceString(dataString, "\\\\", "\n");
                result = result + dataString + "\n";
                newLineFlag = true;
            }
        }
    }

    return result;
}
