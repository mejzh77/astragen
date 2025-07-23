package models

import "github.com/gin-gonic/gin"

func (p *Project) ToDetailedAPI() gin.H {
	return gin.H{
		"id":        p.ID,
		"name":      p.Name,
		"type":      "project",
		"systems":   p.SystemsToDetailedAPI(),
		"createdAt": p.CreatedAt,
		"updatedAt": p.UpdatedAt,
	}
}

func (p *Project) SystemsToDetailedAPI() []gin.H {
	var systems []gin.H
	for _, s := range p.Systems {
		systems = append(systems, s.ToDetailedAPI())
	}
	return systems
}

func (s *System) ToDetailedAPI() gin.H {
	return gin.H{
		"id":        s.ID,
		"name":      s.Name,
		"type":      "system",
		"projectId": s.ProjectID,
		"nodes":     s.NodesToDetailedAPI(),
		"products":  s.ProductsToDetailedAPI(),
	}
}

func (s *System) NodesToDetailedAPI() []gin.H {
	var nodes []gin.H
	for _, n := range s.Nodes {
		nodes = append(nodes, n.ToDetailedAPI())
	}
	return nodes
}

// Для System
func (s *System) ProductsToDetailedAPI() []gin.H {
	var products []gin.H
	for _, p := range s.Products {
		products = append(products, gin.H{
			"id":        p.ID,
			"name":      p.Name,
			"systemId":  p.SystemID,
			"createdAt": p.CreatedAt,
			"updatedAt": p.UpdatedAt,
		})
	}
	return products
}

// Для Node
func (n *Node) ToDetailedAPI() gin.H {
	return gin.H{
		"id":        n.ID,
		"name":      n.Name,
		"type":      "node",
		"systemId":  n.SystemID,
		"createdAt": n.CreatedAt,
		"updatedAt": n.UpdatedAt,
	}
}

// Для Product
func (p *Product) ToDetailedAPI() gin.H {
	return gin.H{
		"id":        p.ID,
		"name":      p.Name,
		"type":      "product",
		"systemId":  p.SystemID,
		"createdAt": p.CreatedAt,
		"updatedAt": p.UpdatedAt,
	}
}

// Для FunctionBlock
func (fb *FunctionBlock) ToDetailedAPI() gin.H {
	return gin.H{
		"id":        fb.ID,
		"tag":       fb.Tag,
		"call":      fb.Call,
		"omx":       fb.OMX,
		"opc":       fb.OPC,
		"type":      "functionblock",
		"cdsType":   fb.CdsType,
		"variables": fb.VariablesToDetailedAPI(),
	}
}

func (fb *FunctionBlock) VariablesToDetailedAPI() []gin.H {
	var vars []gin.H
	for _, v := range fb.Variables {
		vars = append(vars, gin.H{
			"id":        v.ID,
			"direction": v.Direction,
			"signalTag": v.SignalTag,
			"funcAttr":  v.FuncAttr,
			"fbId":      v.FBID,
			"createdAt": v.CreatedAt,
			"updatedAt": v.UpdatedAt,
		})
	}
	return vars
}

// Для Project
func (p *Project) ToAPI() gin.H {
	return gin.H{
		"id":      p.ID,
		"name":    p.Name,
		"type":    "project",
		"systems": p.SystemsToAPI(),
	}
}

// Для System
func (s *System) ToAPI() gin.H {
	return gin.H{
		"id":        s.ID,
		"name":      s.Name,
		"type":      "system",
		"projectId": s.ProjectID,
		"nodes":     s.NodesToAPI(),
		"products":  s.ProductsToAPI(),
	}
}

// Для Node
func (n *Node) ToAPI() gin.H {
	return gin.H{
		"id":             n.ID,
		"name":           n.Name,
		"type":           "node",
		"systemId":       n.SystemID,
		"functionBlocks": n.FunctionBlocksToAPI(),
	}
}

// Для FunctionBlock
func (fb *FunctionBlock) ToAPI() gin.H {
	return gin.H{
		"id":        fb.ID,
		"name":      fb.Tag,
		"type":      "functionblock",
		"system":    fb.System,
		"variables": fb.VariablesToAPI(),
	}
}

func (p *Project) SystemsToAPI() []gin.H {
	var systems []gin.H
	for _, s := range p.Systems {
		systems = append(systems, s.ToAPI())
	}
	return systems
}

func (s *System) ProductsToAPI() []gin.H {
	var products []gin.H
	for _, p := range s.Products {
		products = append(products, gin.H{
			"id":        p.ID,
			"name":      p.Name,
			"type":      "product",
			"systemId":  p.SystemID,
			"createdAt": p.CreatedAt,
			"signals":   p.SignalsToAPI(),
		})
	}
	return products
}
func (p *Product) SignalsToAPI() []gin.H {
	var ss []gin.H
	for _, s := range p.Signals {
		ss = append(ss, s.ToAPI())
	}
	return ss
}

// Для FunctionBlock
func (s *Signal) ToAPI() gin.H {
	return gin.H{
		"id":     s.ID,
		"name":   s.Tag,
		"type":   "signal",
		"system": s.System,
	}
}
func (n *Node) FunctionBlocksToAPI() []gin.H {
	var fbs []gin.H
	for _, fb := range n.FunctionBlocks {
		if !fb.Primary {
			fbs = append(fbs, fb.ToAPI())
		}

	}
	return fbs
}

func (fb *FunctionBlock) VariablesToAPI() []gin.H {
	var vars []gin.H
	for _, v := range fb.Variables {
		vars = append(vars, gin.H{
			"id":        v.ID,
			"type":      "variable",
			"direction": v.Direction,
			"signalTag": v.SignalTag,
			"funcAttr":  v.FuncAttr,
			"fbId":      v.FBID,
		})
	}
	return vars
}

func (s *System) NodesToAPI() []gin.H {
	var nodes []gin.H
	for _, n := range s.Nodes {
		nodes = append(nodes, n.ToAPI())
	}
	return nodes
}
