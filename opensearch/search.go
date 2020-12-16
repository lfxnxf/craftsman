package opensearch

import (
	"context"
	"fmt"
	"github.com/SkyAPM/go2sky"
	"github.com/SkyAPM/go2sky/reporter/grpc/common"
	"github.com/lfxnxf/craftsman/tracing"
	"net/http"
	"strings"
)

type SearchArgs struct {
	//搜索主体
	Query string `ArgName:"query"`
	//要查询的应用名
	Index_name string `ArgName:"index_name"`
	//[可以通过此参数获取本次查询需要的字段内容]
	Fetch_fields string `ArgName:"fetch_fields"`
	//[指定要使用的查询分析规则]
	Qp string `ArgName:"qp"`
	//[关闭已生效的查询分析功能]
	Disable string `ArgName:"disable"`
	//[设置粗排表达式名字]
	First_formula_name string `ArgName:"first_formula_name"`
	//[设置精排表达式名字]
	Formula_name string `ArgName:"formula_name"`
	//[动态摘要的配置]
	Summary string `ArgName:"summary"`
}

//搜索
//系统提供了丰富的搜索语法以满足用户各种场景下的搜索需求
func (this *Client) Search(ctx context.Context, args SearchArgs, resp interface{}) error {
	span, err := this.tracer.CreateExitSpan(ctx, "OpenSearch", "OpenSearch:"+args.Index_name, func(header string) error {
		return nil
	})
	if err != nil {
		this.logger.ErrorT(ctx, "opensearch tracer error", "err", err.Error())
		return this.InvokeByAnyMethod(http.MethodGet, "", "/search", args, resp)
	}

	span.Tag(tracing.TagOpenSearchType, args.Index_name)
	span.SetSpanLayer(common.SpanLayer_Database)
	err = this.InvokeByAnyMethod(http.MethodGet, "", "/search", args, resp)
	if err != nil {
		this.logger.ErrorT(ctx, "opensearch error", "err", err.Error(), "args", args)
		span.Tag(go2sky.TagStatusCode, "err")
	}
	span.End()

	return err
}

func (this *Client) FormatSearch(ctx context.Context, request SearchRequest, searchResp *SearchResp) error {
	args := SearchArgs{
		Index_name:         this.indexName,
		Query:              request.buildQueryClauses(),
		Fetch_fields:       request.buildFetchFieldsClauses(),
		First_formula_name: "",
		Formula_name:       "",
	}

	err := this.Search(ctx, args, searchResp)
	this.logger.InfoT(ctx, "openearch log", "request", request.String(), "resp status", searchResp.Status)
	return err
}

type SearchRequest struct {
	FetchFields []string
	Start       int
	Hits        int
	Kvpairs     string
	Query       string
	Filter      string
	SortFields  SortFields
}

type SortFields []SortField

type SortField struct {
	Field string
	Order string // INCREASE | DECREASE
}

func (sf *SortField) String() string {
	return fmt.Sprintf("%s:%s", sf.Field, sf.Order)
}

func (sfs SortFields) String() string {
	var arr []string
	for _, sf := range sfs {
		arr = append(arr, sf.String())
	}
	return strings.Join(arr, ";")
}

func (req *SearchRequest) buildFetchFieldsClauses() string {
	if len(req.FetchFields) > 0 {
		return strings.Join(req.FetchFields, ";")
	}
	return ""
}

func (req *SearchRequest) buildQueryClauses() string {
	clauses := []string{
		req.defaultConfigClause(),
		req.defaultQueryClause(),
		req.defaultSortClause(),
		req.defaultFilterClause(),
		req.defaultKvpairsClause(),
	}
	sb := strings.Builder{}
	for _, clause := range clauses {
		if len(clause) > 0 {
			sb.WriteString("&&" + clause)
		}
	}
	//fmt.Println("search-query:", strings.TrimLeft(sb.String(), "&&"))
	return strings.TrimLeft(sb.String(), "&&")
}

func (req *SearchRequest) defaultConfigClause() string {
	sb := strings.Builder{}
	sb.WriteString("config=")
	sb.WriteString(fmt.Sprintf("start:%d,", req.Start))
	sb.WriteString(fmt.Sprintf("hit:%d,", req.Hits))
	sb.WriteString("format:fulljson")
	return sb.String()
}

func (req *SearchRequest) defaultQueryClause() string {
	return "query=" + req.Query
}

func (req *SearchRequest) defaultSortClause() string {
	if len(req.SortFields) > 0 && req.SortFields.String() != ":" {
		sb := strings.Builder{}
		sb.WriteString("sort=")
		for _, sortField := range req.SortFields {
			sortStr := sortField.Field
			switch sortField.Order {
			case "INCREASE", "increase", "asc":
				sortStr = "+" + sortStr
			default:
				sortStr = "-" + sortStr
			}
			sb.WriteString(sortStr + ";")
		}
		return strings.TrimRight(sb.String(), ";")
	}
	return ""
}

func (req *SearchRequest) defaultFilterClause() string {
	if len(req.Filter) > 0 {
		return "filter=" + req.Filter
	}
	return ""
}

func (req *SearchRequest) defaultKvpairsClause() string {
	if len(req.Kvpairs) > 0 {
		return "kvpairs=" + req.Kvpairs
	}
	return ""
}

func (req *SearchRequest) String() string {
	return fmt.Sprintf(`{"fetch_fields": %#v, "start": %d, "hits": %d, "kvpairs": %#v, "query": %#v, "filter": %#v, "sort_fields": %#v}`,
		req.FetchFields, req.Start, req.Hits, req.Kvpairs, req.Query, req.Filter, req.SortFields.String())
}

type SearchResp struct {
	Status    string   `json:"status"`
	RequestId string   `json:"request_id"`
	Res       *Result  `json:"result"`
	Errors    []*Error `json:"errors"`
}

type Result struct {
	SearchTime float64 `json:"searchtime"`
	Total      int     `json:"total"`
	Num        int     `json:"num"`
	ViewTotal  int     `json:"viewtotal"`
	Items      []*Item `json:"items"`
}

type Item struct {
	Fields         map[string]string `json:"fields"`
	SortExprValues []string          `json:"sortExprValues"`
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
