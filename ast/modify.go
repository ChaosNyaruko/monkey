package ast

type Modifer func(Node) Node

func Modify(node Node, f Modifer) Node {
	switch node := node.(type) {
	case *Program:
		for i, s := range node.Statements {
			node.Statements[i] = Modify(s, f).(Statement)
		}
		return node

	case *ExpressionStatement:
		node.Expression = Modify(node.Expression, f).(Expression)
		return node
	case *InfixExpression:
		node.Lhs = Modify(node.Lhs, f).(Expression)
		node.Rhs = Modify(node.Rhs, f).(Expression)
		return node
	case *PrefixExpression:
		node.Rhs = Modify(node.Rhs, f).(Expression)
		return node
	case *IndexExpression:
		node.Left = Modify(node.Left, f).(Expression)
		node.Index = Modify(node.Index, f).(Expression)
		return node
	case *BlockStatement:
		for i, s := range node.Statements {
			node.Statements[i] = Modify(s, f).(Statement)
		}
		return node
	case *IfExpression:
		node.Condition = Modify(node.Condition, f).(Expression)
		node.If = Modify(node.If, f).(*BlockStatement)
		node.Else = Modify(node.Else, f).(*BlockStatement)
		return node
	case *ReturnStatement:
		node.ReturnValue = Modify(node.ReturnValue, f).(Expression)
		return node
	case *LetStatement:
		node.Value = Modify(node.Value, f).(Expression)
		return node
	case *FunctionLiteral:
		for i, p := range node.Parameters {
			node.Parameters[i] = Modify(p, f).(*Identifier)
		}
		node.Body = Modify(node.Body, f).(*BlockStatement)
		return node
	case *ArrayLiteral:
		for i, e := range node.Elements {
			node.Elements[i] = Modify(e, f).(Expression)
		}
		return node

	case *HashLiteral:
		newPairs := make(map[Expression]Expression)
		for k, v := range node.Pairs {
			nk := Modify(k, f).(Expression)
			nv := Modify(v, f).(Expression)
			newPairs[nk] = nv
		}
		node.Pairs = newPairs
		return node
	}

	return f(node)
}
