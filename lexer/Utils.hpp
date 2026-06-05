#pragma once

#include <string>
#include "TokenType.hpp"

// Função auxiliar para converter o enum TokenType em texto.
// Isso facilita a exibição dos tokens no terminal.
std::string tokenTypeToString(TokenType type);
