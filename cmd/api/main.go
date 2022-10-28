package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/hyperjumptech/grule-rule-engine/ast"
	"github.com/hyperjumptech/grule-rule-engine/builder"
	"github.com/hyperjumptech/grule-rule-engine/engine"
	"github.com/hyperjumptech/grule-rule-engine/pkg"
)

type TypedJson struct {
	ID string
}

type ObjectResult struct {
	Result string
}

type Template struct {
	State string
	ID    string
}

type RequesBody struct {
	Payment string
	Amount  float32
	Result  bool
}

type Fact struct {
	NetAmount float32
	Distance  int32
	Duration  int32
	Result    bool
}

func main() {
	http.HandleFunc("/generic-body", func(w http.ResponseWriter, r *http.Request) {

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			fmt.Fprintf(w, "err =  %s\n", err)
		}

		fmt.Fprintf(w, "Body =  %s\n", body)
		fmt.Fprintf(w, "End request to  %s\n", r.URL.Path)
	})

	http.HandleFunc("/tiped-body", func(w http.ResponseWriter, r *http.Request) {

		var requestJson TypedJson

		err := json.NewDecoder(r.Body).Decode(&requestJson)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		fmt.Fprintf(w, "Body =  %s\n", requestJson)

		json2 := &TypedJson{ID: "987654"}
		b, err := json.Marshal(json2)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Fprintf(w, "json =  %s\n", string(b))

		fmt.Fprintf(w, "End request to  %s\n", r.URL.Path)
	})

	http.HandleFunc("/teste", func(w http.ResponseWriter, r *http.Request) {

		rule := `rule CheckIfJSONStringWorks {
					when R.ID != nil && R.ID == "12345" 
					then R.ID = "PERFECT";
				}`

		var requestJson TypedJson

		err := json.NewDecoder(r.Body).Decode(&requestJson)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		dataContext := ast.NewDataContext()
		error := dataContext.Add("R", requestJson)
		if error != nil {
			fmt.Fprintf(w, "Error =  %s\n", err)
		}

		lib := ast.NewKnowledgeLibrary()
		rb := builder.NewRuleBuilder(lib)
		err2 := rb.BuildRuleFromResource("TestJSONSimple", "0.0.1", pkg.NewBytesResource([]byte(rule)))
		if err2 != nil {
			fmt.Fprintf(w, "Error =  %s\n", err2)
		}

		kb := lib.NewKnowledgeBaseInstance("TestJSONSimple", "0.0.1")

		e := engine.NewGruleEngine()
		ruleEntries, err := e.FetchMatchingRules(dataContext, kb)
		if err != nil {
			fmt.Fprintf(w, "Error =  %s\n", err2)
		}

		fmt.Fprintf(w, "Error =  %s\n", ruleEntries[0].RuleName)

		fmt.Fprintf(w, "End request to  %s\n", r.URL.Path)
	})

	http.HandleFunc("/novo-teste", func(w http.ResponseWriter, r *http.Request) {
		oresult := &ObjectResult{
			Result: "NoResult",
		}

		dataContext := ast.NewDataContext()
		dataContext.Add("R", oresult)

		_ = dataContext.AddJSON("str", []byte(`"A String"`))

		rule := `
		rule CheckIfJSONStringWorks {
			when R.Result == "NoResult" && str.ToUpper() == "A STRING" 
			then R.Result = "PERFECT";
		}`

		// Prepare knowledgebase library and load it with our rule.
		lib := ast.NewKnowledgeLibrary()
		rb := builder.NewRuleBuilder(lib)
		err := rb.BuildRuleFromResource("TestJSONSimple", "0.0.1", pkg.NewBytesResource([]byte(rule)))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		eng1 := &engine.GruleEngine{MaxCycle: 5}
		kb := lib.NewKnowledgeBaseInstance("TestJSONSimple", "0.0.1")
		err = eng1.Execute(dataContext, kb)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		fmt.Fprintf(w, "End request oresult  %s\n", oresult)
	})

	http.HandleFunc("/teste-json", func(w http.ResponseWriter, r *http.Request) {
		/* JSON EXEMPLO:
		{
			"payment": "123",
			"payment_method": "bolbradesco",
			"amount": 100,
			"tax": 1,
			"site": "MELI"
		}
		*/
		template := &Template{
			State: "No Result",
		}
		dataContext := ast.NewDataContext()
		err := dataContext.Add("Result", template)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		body, _ := ioutil.ReadAll(r.Body)

		err = dataContext.AddJSON("json", body)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		rule := `
	rule CheckIfJSONIntWorks {
		when 
			Result.State == "No Result" && 
			json.payment == "123" &&
			json.payment == "123" &&
			json.payment_method == "bolbradesco" &&
			json.amount > 10 &&
			json.amount < 1000 &&
			json.tax <= 10 &&
			json.tax >= 1 &&
			json.site != "ASD"
		then
			Result.State = "FOUND";
			Result.ID = "456";
	}`

		// Prepare knowledgebase library and load it with our rule.
		lib := ast.NewKnowledgeLibrary()
		rb := builder.NewRuleBuilder(lib)
		err = rb.BuildRuleFromResource("TestJSONBitComplex", "0.0.1", pkg.NewBytesResource([]byte(rule)))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		eng1 := &engine.GruleEngine{MaxCycle: 5}
		kb := lib.NewKnowledgeBaseInstance("TestJSONBitComplex", "0.0.1")
		err = eng1.Execute(dataContext, kb)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		b, _ := json.Marshal(template)
		fmt.Fprintf(w, "%s\n", string(b))
	})

	http.HandleFunc("/teste-regras-repetidas", func(w http.ResponseWriter, r *http.Request) {

		const rulesWithDiffSalience = `
			rule Rule1 "Rule 1" salience 5 {
			when
			(R.Payment != "" && R.Amount >= 5000) && (R.Result == false)
			Then
			R.Result=true;
			}

			rule Rule2 "Rule 2" salience 6 {
			when
			(R.Payment != "" && R.Amount > 6000) && (R.Result == false)
			Then
			R.Result=true;
			}

			rule Rule3 "Rule 3" salience 7 {
			when
			(R.Payment != "" && R.Amount > 7000) && (R.Result == false)
			Then
			R.Result=true;
			}

			rule Rule4 "Rule 4" salience 8 {
			when
			(R.Payment != "" && R.Amount > 8000) && (R.Result == false)
			Then
			R.Result=true;
			}
			
			rule Rule5 "Rule 5" salience 9 {
			when
			(R.Payment != "" && R.Amount == 5000) && (R.Result == false)
			Then
			R.Result=true;
			}`

		var requestBody RequesBody

		err := json.NewDecoder(r.Body).Decode(&requestBody)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		dctx := ast.NewDataContext()
		err = dctx.Add("R", requestBody)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		lib := ast.NewKnowledgeLibrary()
		rb := builder.NewRuleBuilder(lib)
		err = rb.BuildRuleFromResource("rules_test", "0.1.1", pkg.NewBytesResource([]byte(rulesWithDiffSalience)))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		kb := lib.NewKnowledgeBaseInstance("rules_test", "0.1.1")

		e := engine.NewGruleEngine()
		ruleEntries, err := e.FetchMatchingRules(dctx, kb)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		fmt.Fprintf(w, "Number of ruleEntries  %d\n", len(ruleEntries))
		for position, rule := range ruleEntries {
			fmt.Fprintf(w, "Rule Name = %s AND Position = %d AND Salience = %d\n", rule.RuleName, position, rule.Salience)
		}
	})

	http.ListenAndServe(":8080", nil)
}
