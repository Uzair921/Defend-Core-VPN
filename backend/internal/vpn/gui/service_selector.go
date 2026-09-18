package gui

import (
"fmt"

"fyne.io/fyne/v2"
"fyne.io/fyne/v2/container"
"fyne.io/fyne/v2/theme"
"fyne.io/fyne/v2/widget"
)

// ServiceSelector renders a list of VPN services as selectable cards.
type ServiceSelector struct {
services   []MyService
selected   string
onSelect   func(MyService)
container  *fyne.Container
}

// NewServiceSelector creates a new service selector widget.
func NewServiceSelector(onSelect func(MyService)) *ServiceSelector {
s := &ServiceSelector{
onSelect:  onSelect,
container: container.NewVBox(),
}
return s
}

// SetServices updates the list of services and rebuilds the UI.
func (s *ServiceSelector) SetServices(services []MyService) {
s.services = services
s.rebuild()
}

// Container returns the underlying Fyne container.
func (s *ServiceSelector) Container() fyne.CanvasObject {
return s.container
}

// Selected returns the currently selected service, if any.
func (s *ServiceSelector) Selected() *MyService {
for i := range s.services {
if s.services[i].ID == s.selected {
return &s.services[i]
}
}
return nil
}

func (s *ServiceSelector) rebuild() {
s.container.RemoveAll()

if len(s.services) == 0 {
s.container.Add(widget.NewLabelWithStyle(
"No VPN services assigned to your organization.",
fyne.TextAlignCenter,
fyne.TextStyle{Italic: true},
))
s.container.Refresh()
return
}

for _, svc := range s.services {
s.container.Add(s.makeCard(svc))
}
s.container.Refresh()
}

func (s *ServiceSelector) makeCard(svc MyService) fyne.CanvasObject {
icon := widget.NewLabel(svc.TypeIcon)
icon.TextStyle = fyne.TextStyle{Bold: true}

name := widget.NewLabelWithStyle(svc.Name, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
typeName := widget.NewLabel(svc.TypeName)
typeName.TextStyle = fyne.TextStyle{Italic: true}

meta := widget.NewLabel(fmt.Sprintf("Max clients: %d  ·  Subnet: %s", svc.MaxClients, svc.Subnet))

header := container.NewHBox(icon, container.NewVBox(name, typeName))

btn := widget.NewButtonWithIcon("Connect", theme.MediaPlayIcon(), func() {
s.selected = svc.ID
if s.onSelect != nil {
s.onSelect(svc)
}
})
btn.Importance = widget.HighImportance

card := container.NewVBox(
header,
meta,
container.NewHBox(btn),
)

return widget.NewCard("", "", card)
}
