#include "Utils.hpp"

std::string tokenTypeToString(TokenType type) {

    switch(type) {

    case TokenType::T_PACKAGE:        return "T_PACKAGE";
    case TokenType::T_IMPORT:         return "T_IMPORT";
    case TokenType::T_FUNC:           return "T_FUNC";
    case TokenType::T_VAR:            return "T_VAR";
    case TokenType::T_INT:            return "T_INT";
    case TokenType::T_FLOAT:          return "T_FLOAT";
    case TokenType::T_BOOL:           return "T_BOOL";
    case TokenType::T_STRING:         return "T_STRING";
    case TokenType::T_IF:             return "T_IF";
    case TokenType::T_ELSE:           return "T_ELSE";
    case TokenType::T_FOR:            return "T_FOR";
    case TokenType::T_TRUE:           return "T_TRUE";
    case TokenType::T_FALSE:          return "T_FALSE";
    case TokenType::T_ID:             return "T_ID";
    case TokenType::T_NUM:            return "T_NUM";
    case TokenType::T_FLOAT_NUM:      return "T_FLOAT_NUM";
    case TokenType::T_STRING_LITERAL: return "T_STRING_LITERAL";
    case TokenType::T_ASSIGN:         return "T_ASSIGN";
    case TokenType::T_DECLARE_ASSIGN: return "T_DECLARE_ASSIGN";
    case TokenType::T_EQ:             return "T_EQ";
    case TokenType::T_PLUS:           return "T_PLUS";
    case TokenType::T_MINUS:          return "T_MINUS";
    case TokenType::T_MULT:           return "T_MULT";
    case TokenType::T_DIV:            return "T_DIV";
    case TokenType::T_LT:             return "T_LT";
    case TokenType::T_GT:             return "T_GT";
    case TokenType::T_LPAREN:         return "T_LPAREN";
    case TokenType::T_RPAREN:         return "T_RPAREN";
    case TokenType::T_LBRACE:         return "T_LBRACE";
    case TokenType::T_RBRACE:         return "T_RBRACE";
    case TokenType::T_SEMICOLON:      return "T_SEMICOLON";
    case TokenType::T_COLON:          return "T_COLON";
    case TokenType::T_EOF:            return "T_EOF";
    default:                          return "UNKNOWN";
    }
}
