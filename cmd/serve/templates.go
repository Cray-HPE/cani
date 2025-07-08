package serve

import (
	"html/template"
	"log"
	"net/http"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

type TemplateManager struct {
	funcMap template.FuncMap
}

func NewTemplateManager() *TemplateManager {
	funcMap := template.FuncMap{
		"devicetypeTypeEquals": func(t devicetypes.Type, s string) bool {
			return string(t) == s
		},
	}

	return &TemplateManager{
		funcMap: funcMap,
	}
}

func (tm *TemplateManager) Render(w http.ResponseWriter, templateName string, data interface{}) {
	templates := map[string][]string{
		"dashboard":    {"templates/base.html", "templates/dashboard.html"},
		"device":       {"templates/base.html", "templates/device.html"},
		"devices":      {"templates/base.html", "templates/devices.html"},
		"devicetypes":  {"templates/base.html", "templates/devicetypes.html"},
		"devicetype":   {"templates/base.html", "templates/devicetype.html"},
		"racks":        {"templates/base.html", "templates/racks.html"},
		"staged":       {"templates/base.html", "templates/staged.html"},
		"production":   {"templates/base.html", "templates/production.html"},
		"maintenance":  {"templates/base.html", "templates/maintenance.html"},
		"decommission": {"templates/base.html", "templates/decommission.html"},
	}

	templateFiles, ok := templates[templateName]
	if !ok {
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return
	}

	tpl := template.New("base").Funcs(tm.funcMap)
	tpl = template.Must(tpl.ParseFiles(templateFiles...))

	if err := tpl.ExecuteTemplate(w, "base", data); err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
