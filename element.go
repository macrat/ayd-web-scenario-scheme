package main

import (
	"context"
	"fmt"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	"github.com/yuin/gopher-lua"
)

type Element struct {
	query string
	ids   []cdp.NodeID
	tab   *Tab
}

func NewElement(L *lua.LState, t *Tab, query string) Element {
	var ids []cdp.NodeID
	t.RunSelector(L, query, chromedp.NodeIDs(query, &ids, chromedp.ByQuery))

	return Element{
		query: query,
		ids:   ids,
		tab:   t,
	}
}

func (e Element) ToLua(L *lua.LState) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = e
	L.SetMetatable(ud, L.GetTypeMetatable("element"))
	return ud
}

func CheckElement(L *lua.LState) Element {
	if ud, ok := L.Get(1).(*lua.LUserData); ok {
		if e, ok := ud.Value.(Element); ok {
			return e
		}
	}

	L.ArgError(1, "element expected. perhaps you call it like tab().xxx() instead of tab():xxx().")
	return Element{}
}

func (e Element) Select(L *lua.LState, query string) Element {
	var nodes []*cdp.Node
	e.tab.Run(L, chromedp.Nodes(e.ids, &nodes, chromedp.ByNodeID))

	var ids []cdp.NodeID
	e.tab.Run(L, chromedp.NodeIDs(query, &ids, chromedp.ByQuery, chromedp.FromNode(nodes[0])))

	return Element{
		query: query,
		ids:   ids,
		tab:   e.tab,
	}
}

func (e Element) SelectAll(L *lua.LState, query string) ElementsArray {
	var nodes []*cdp.Node
	e.tab.RunSelector(L, query, chromedp.Nodes(e.ids, &nodes, chromedp.ByNodeID))

	var es ElementsArray
	for _, node := range nodes {
		var ids []cdp.NodeID
		e.tab.RunSelector(L, query, chromedp.NodeIDs(
			query,
			&ids,
			chromedp.ByQueryAll,
			chromedp.FromNode(node),
			chromedp.AtLeast(0),
		))
		for _, id := range ids {
			es = append(es, Element{
				query: query,
				ids:   []cdp.NodeID{id},
				tab:   e.tab,
			})
		}
	}
	return es
}

func (e Element) SendKeys(L *lua.LState) {
	text := L.ToString(2)
	e.tab.Run(L, chromedp.SendKeys(e.ids, text, chromedp.ByNodeID))
}

func (e Element) SetValue(L *lua.LState) {
	value := L.ToString(2)
	e.tab.Run(L, chromedp.SetValue(e.ids, value, chromedp.ByNodeID))
}

func (e Element) Click(L *lua.LState) {
	e.tab.Run(L, chromedp.Click(e.ids, chromedp.ByNodeID))
}

func (e Element) Submit(L *lua.LState) {
	e.tab.Run(L, chromedp.Submit(e.ids, chromedp.ByNodeID))
}

func (e Element) Focus(L *lua.LState) {
	e.tab.Run(L, chromedp.Focus(e.ids, chromedp.ByNodeID))
}

func (e Element) Blur(L *lua.LState) {
	e.tab.Run(L, chromedp.Blur(e.ids, chromedp.ByNodeID))
}

func (e Element) Screenshot(L *lua.LState) {
	name := L.ToString(2)

	var buf []byte
	e.tab.Run(L, chromedp.Screenshot(e.ids, &buf, chromedp.ByNodeID))
	e.tab.Save(L, name, ".jpg", buf)
}

func (e Element) GetText(L *lua.LState) int {
	var text string
	e.tab.Run(L, chromedp.Text(e.ids, &text, chromedp.ByNodeID))
	L.Push(lua.LString(text))
	return 1
}

func (e Element) GetInnerHTML(L *lua.LState) int {
	var html string
	e.tab.Run(L, chromedp.InnerHTML(e.ids, &html, chromedp.ByNodeID))
	L.Push(lua.LString(html))
	return 1
}

func (e Element) GetOuterHTML(L *lua.LState) int {
	var html string
	e.tab.Run(L, chromedp.OuterHTML(e.ids, &html, chromedp.ByNodeID))
	L.Push(lua.LString(html))
	return 1
}

func (e Element) GetValue(L *lua.LState) int {
	var value string
	e.tab.Run(L, chromedp.Value(e.ids, &value, chromedp.ByNodeID))
	L.Push(lua.LString(value))
	return 1
}

func (e Element) GetAttribute(L *lua.LState) int {
	name := L.CheckString(2)

	var value string
	var ok bool
	e.tab.Run(L, chromedp.AttributeValue(e.ids, name, &value, &ok, chromedp.ByNodeID))

	if ok {
		L.Push(lua.LString(value))
		return 1
	} else {
		return 0
	}
}

func RegisterElementType(ctx context.Context, L *lua.LState) {
	fn := func(f func(Element, *lua.LState)) *lua.LFunction {
		return L.NewFunction(func(L *lua.LState) int {
			f(CheckElement(L), L)
			L.Push(L.Get(1))
			return 1
		})
	}

	methods := map[string]*lua.LFunction{
		"all": L.NewFunction(func(L *lua.LState) int {
			e := CheckElement(L)
			query := L.CheckString(2)
			L.Push(e.SelectAll(L, query).ToLua(L))
			return 1
		}),
		"sendKeys":   fn(Element.SendKeys),
		"setValue":   fn(Element.SetValue),
		"click":      fn(Element.Click),
		"submit":     fn(Element.Submit),
		"focus":      fn(Element.Focus),
		"blur":       fn(Element.Blur),
		"screenshot": fn(Element.Screenshot),
	}

	getters := map[string]func(Element, *lua.LState) int{
		"text":      Element.GetText,
		"innerHTML": Element.GetInnerHTML,
		"outerHTML": Element.GetOuterHTML,
		"value":     Element.GetValue,
	}

	query := L.SetFuncs(L.NewTypeMetatable("element"), map[string]lua.LGFunction{
		"__call": func(L *lua.LState) int {
			e := CheckElement(L)
			query := L.CheckString(2)
			L.Push(e.Select(L, query).ToLua(L))
			return 1
		},
		"__index": func(L *lua.LState) int {
			name := L.CheckString(2)

			if f, ok := getters[name]; ok {
				return f(CheckElement(L), L)
			} else if f, ok := methods[name]; ok {
				L.Push(f)
				return 1
			} else {
				return CheckElement(L).GetAttribute(L)
			}
		},
		"__tostring": func(L *lua.LState) int {
			e := CheckElement(L)
			L.Push(lua.LString(fmt.Sprintf("{%s}", e.query)))
			return 1
		},
	})
	L.SetGlobal("element", query)
}
