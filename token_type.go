package pinky

type TokenType int

const (
	TokenLPAREN TokenType = iota
	TokenRPAREN
	TokenLCURLY
	TokenRCURLY
	TokenLSQUAR
	TokenRSQUAR
	TokenCOMMA
	TokenDOT
	TokenPLUS
	TokenMINUS
	TokenSTAR
	TokenSLASH
	TokenCARET
	TokenMOD
	TokenCOLON
	TokenSEMICOLON
	TokenQUESTION
	TokenNOT
	TokenGT
	TokenLT
	TokenEQ
	TokenGE
	TokenLE
	TokenNE
	TokenEQEQ
	TokenASSIGN
	TokenGTGT
	TokenLTLT
	TokenIDENTIFIER
	TokenSTRING
	TokenINTEGER
	TokenFLOAT
	TokenIF
	TokenTHEN
	TokenELSE
	TokenTRUE
	TokenFALSE
	TokenAND
	TokenOR
	TokenLOCAL
	TokenWHILE
	TokenDO
	TokenFOR
	TokenFUNC
	TokenNULL
	TokenEND
	TokenPRINT
	TokenPRINTLN
	TokenRET
)

func tokenDebugName(kind TokenType) string {
	switch kind {
	case TokenLPAREN:
		return "TOK_LPAREN"
	case TokenRPAREN:
		return "TOK_RPAREN"
	case TokenLCURLY:
		return "TOK_LCURLY"
	case TokenRCURLY:
		return "TOK_RCURLY"
	case TokenLSQUAR:
		return "TOK_LSQUAR"
	case TokenRSQUAR:
		return "TOK_RSQUAR"
	case TokenCOMMA:
		return "TOK_COMMA"
	case TokenDOT:
		return "TOK_DOT"
	case TokenPLUS:
		return "TOK_PLUS"
	case TokenMINUS:
		return "TOK_MINUS"
	case TokenSTAR:
		return "TOK_STAR"
	case TokenSLASH:
		return "TOK_SLASH"
	case TokenCARET:
		return "TOK_CARET"
	case TokenMOD:
		return "TOK_MOD"
	case TokenCOLON:
		return "TOK_COLON"
	case TokenSEMICOLON:
		return "TOK_SEMICOLON"
	case TokenQUESTION:
		return "TOK_QUESTION"
	case TokenNOT:
		return "TOK_NOT"
	case TokenGT:
		return "TOK_GT"
	case TokenLT:
		return "TOK_LT"
	case TokenEQ:
		return "TOK_EQ"
	case TokenGE:
		return "TOK_GE"
	case TokenLE:
		return "TOK_LE"
	case TokenNE:
		return "TOK_NE"
	case TokenEQEQ:
		return "TOK_EQEQ"
	case TokenASSIGN:
		return "TOK_ASSIGN"
	case TokenGTGT:
		return "TOK_GTGT"
	case TokenLTLT:
		return "TOK_LTLT"
	case TokenIDENTIFIER:
		return "TOK_IDENTIFIER"
	case TokenSTRING:
		return "TOK_STRING"
	case TokenINTEGER:
		return "TOK_INTEGER"
	case TokenFLOAT:
		return "TOK_FLOAT"
	case TokenIF:
		return "TOK_IF"
	case TokenTHEN:
		return "TOK_THEN"
	case TokenELSE:
		return "TOK_ELSE"
	case TokenTRUE:
		return "TOK_TRUE"
	case TokenFALSE:
		return "TOK_FALSE"
	case TokenAND:
		return "TOK_AND"
	case TokenOR:
		return "TOK_OR"
	case TokenLOCAL:
		return "TOK_LOCAL"
	case TokenWHILE:
		return "TOK_WHILE"
	case TokenDO:
		return "TOK_DO"
	case TokenFOR:
		return "TOK_FOR"
	case TokenFUNC:
		return "TOK_FUNC"
	case TokenNULL:
		return "TOK_NULL"
	case TokenEND:
		return "TOK_END"
	case TokenPRINT:
		return "TOK_PRINT"
	case TokenPRINTLN:
		return "TOK_PRINTLN"
	case TokenRET:
		return "TOK_RET"
	default:
		return "TOK_UNKNOWN"
	}
}
