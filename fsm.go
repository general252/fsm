package fsm

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

type State string

type Event string

type TransitionHandler func(from State, e Event, to State) error

type eKey struct {
	From  State
	Event Event
}

type Transition struct {
	From   State
	Event  Event
	To     State
	Handle TransitionHandler
}

type StateMachine struct {
	current     State
	transitions map[eKey]*Transition
	mutex       sync.Mutex
}

func NewStateMachine(current State) *StateMachine {
	return &StateMachine{current: current}
}

func (fm *StateMachine) CurrentState() State {
	return fm.current
}

func (fm *StateMachine) Trigger(event Event) error {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	if trans, ok := fm.transitions[eKey{fm.current, event}]; ok {
		if err := trans.Handle(fm.current, event, trans.To); err != nil {
			return err
		}
		fm.current = trans.To
		return nil
	}

	return fmt.Errorf("state, event: [%v, %v] undefined", fm.current, event)
}

func (fm *StateMachine) AddTransitions(transitions ...*Transition) error {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	for _, transition := range transitions {
		if err := fm.addTransition(transition); err != nil {
			return err
		}
	}

	return nil
}

func (fm *StateMachine) addTransition(transition *Transition) error {
	var (
		from  = transition.From
		event = transition.Event
	)

	if fm.transitions == nil {
		fm.transitions = make(map[eKey]*Transition)
	}

	if _, ok := fm.transitions[eKey{from, event}]; ok {
		return fmt.Errorf("state, event: [%v, %v] existed", from, event)
	}

	fm.transitions[eKey{from, event}] = transition
	return nil
}

// View
// https://www.mermaidchart.com/play
// http://www.webgraphviz.com/
func (fm *StateMachine) View() (graphViz, flowChart, diagram string) {

	var getSortedTransitionKeys = func(transitions map[eKey]*Transition) []eKey {
		sortedTransitionKeys := make([]eKey, 0)

		for transition := range transitions {
			sortedTransitionKeys = append(sortedTransitionKeys, transition)
		}
		sort.Slice(sortedTransitionKeys, func(i, j int) bool {
			if sortedTransitionKeys[i].From == sortedTransitionKeys[j].From {
				return sortedTransitionKeys[i].Event < sortedTransitionKeys[j].Event
			}
			return sortedTransitionKeys[i].From < sortedTransitionKeys[j].From
		})

		return sortedTransitionKeys
	}

	var getSortedStates = func(transitions map[eKey]*Transition) ([]string, map[string]string) {
		statesToIDMap := make(map[string]string)
		for transition, target := range transitions {
			if _, ok := statesToIDMap[string(transition.From)]; !ok {
				statesToIDMap[string(transition.From)] = ""
			}
			if _, ok := statesToIDMap[string(target.To)]; !ok {
				statesToIDMap[string(target.To)] = ""
			}
		}

		sortedStates := make([]string, 0, len(statesToIDMap))
		for state := range statesToIDMap {
			sortedStates = append(sortedStates, state)
		}
		sort.Strings(sortedStates)

		for i, state := range sortedStates {
			statesToIDMap[state] = fmt.Sprintf("id%d", i)
		}
		return sortedStates, statesToIDMap
	}

	sortedTransitionKeys := getSortedTransitionKeys(fm.transitions)
	sortedStates, statesToIDMap := getSortedStates(fm.transitions)

	var bufFlowChart strings.Builder
	{
		// writeFlowChartGraphType
		bufFlowChart.WriteString("graph LR\n")

		// writeFlowChartStates
		for _, state := range sortedStates {
			bufFlowChart.WriteString(fmt.Sprintf(`    %s[%s]`, statesToIDMap[state], state))
			bufFlowChart.WriteString("\n")
		}
		bufFlowChart.WriteString("\n")

		// writeFlowChartTransitions
		for _, transition := range sortedTransitionKeys {
			target := fm.transitions[transition]
			bufFlowChart.WriteString(fmt.Sprintf(`    %s --> |%s| %s`, statesToIDMap[string(transition.From)], string(transition.Event), statesToIDMap[string(target.To)]))
			bufFlowChart.WriteString("\n")
		}
		bufFlowChart.WriteString("\n")

		// writeFlowChartHighlightCurrent
		const highlightingColor = "#00AA00"
		bufFlowChart.WriteString(fmt.Sprintf(`    style %s fill:%s`, statesToIDMap[string(fm.current)], highlightingColor))
		bufFlowChart.WriteString("\n")
	}

	var bufDiagram strings.Builder
	{
		bufDiagram.WriteString("stateDiagram\n")
		bufDiagram.WriteString(fmt.Sprintln(`    [*] -->`, string(fm.current)))

		for _, k := range sortedTransitionKeys {
			v := fm.transitions[k]
			bufDiagram.WriteString(fmt.Sprintf(`    %s --> %s: %s`, string(k.From), string(v.To), string(k.Event)))
			bufDiagram.WriteString("\n")
		}
	}

	var bufGraphViz strings.Builder
	{
		// writeHeaderLine
		bufGraphViz.WriteString(`digraph fsm {`)
		bufGraphViz.WriteString("\n")

		// writeTransitions
		for _, k := range sortedTransitionKeys {
			v := fm.transitions[k]
			bufGraphViz.WriteString(fmt.Sprintf(`    "%s" -> "%s" [ label = "%s" ];`, string(k.From), string(v.To), string(k.Event)))
			bufGraphViz.WriteString("\n")
		}

		bufGraphViz.WriteString("\n")

		// writeStates
		for _, k := range sortedStates {
			if k == string(fm.current) {
				bufGraphViz.WriteString(fmt.Sprintf(`    "%s" [color = "red"];`, k))
			} else {
				bufGraphViz.WriteString(fmt.Sprintf(`    "%s";`, k))
			}
			bufGraphViz.WriteString("\n")
		}

		// writeFooter
		bufGraphViz.WriteString(fmt.Sprintln("}"))
	}

	return bufGraphViz.String(), bufFlowChart.String(), bufDiagram.String()
}
