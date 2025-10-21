// rattype - C++ implementation of zootype typing test

#include <iostream>
#include <string>

const char* VERSION = "dev";

int main(int argc, char* argv[]) {
    if (argc > 1 && (std::string(argv[1]) == "--version" || std::string(argv[1]) == "-v")) {
        std::cout << "rattype " << VERSION << std::endl;
        return 0;
    }
    std::cout << "Hello from rattype" << std::endl;
    return 0;
}
