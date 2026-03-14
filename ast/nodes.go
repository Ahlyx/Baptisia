package ast

type Node interface {
	NodeType() string
	GetLine() int
}

type Program struct {
	Device *DeviceNode
}

type DeviceNode struct {
	Line     int
	Name     string
	Target   string
	Watchdog string
	Cycle    string
	Consts   *ConstNode
	States   *StatesNode
	Vars     *VarsNode
	Boot     *BootNode
	Inputs   *InputsNode
	Outputs  *OutputsNode
	Safety   *SafetyNode
	Failsafe *FailsafeNode
	Control  *ControlNode
}

func (d *DeviceNode) NodeType() string { return "Device" }
func (d *DeviceNode) GetLine() int     { return d.Line }

type StatesNode struct {
	Line  int
	Names []string
}

func (s *StatesNode) NodeType() string { return "States" }
func (s *StatesNode) GetLine() int     { return s.Line }

type VarsNode struct {
	Line int
	Vars []VarDecl
}

func (v *VarsNode) NodeType() string { return "Vars" }
func (v *VarsNode) GetLine() int     { return v.Line }

type VarDecl struct {
	Line     int
	Volatile bool
	TypeName string
	Name     string
	Value    string
}

func (v *VarDecl) NodeType() string { return "VarDecl" }
func (v *VarDecl) GetLine() int     { return v.Line }

type BootNode struct {
	Line       int
	Statements []Node
}

func (b *BootNode) NodeType() string { return "Boot" }
func (b *BootNode) GetLine() int     { return b.Line }

type InputsNode struct {
	Line   int
	Inputs []SensorAssign
}

func (i *InputsNode) NodeType() string { return "Inputs" }
func (i *InputsNode) GetLine() int     { return i.Line }

type SensorAssign struct {
	Line int
	Name string
	Pin  string
	Min  string
	Max  string
}

func (s *SensorAssign) NodeType() string { return "SensorAssign" }
func (s *SensorAssign) GetLine() int     { return s.Line }

type OutputsNode struct {
	Line    int
	Outputs []ActuatorDecl
}

func (o *OutputsNode) NodeType() string { return "Outputs" }
func (o *OutputsNode) GetLine() int     { return o.Line }

type ActuatorDecl struct {
	Line int
	Name string
	Pin  string
}

func (a *ActuatorDecl) NodeType() string { return "ActuatorDecl" }
func (a *ActuatorDecl) GetLine() int     { return a.Line }

type SafetyNode struct {
	Line       int
	Statements []Node
}

func (s *SafetyNode) NodeType() string { return "Safety" }
func (s *SafetyNode) GetLine() int     { return s.Line }

type FailsafeNode struct {
	Line       int
	Statements []Node
}

func (f *FailsafeNode) NodeType() string { return "Failsafe" }
func (f *FailsafeNode) GetLine() int     { return f.Line }

type ControlNode struct {
	Line       int
	Statements []Node
}

func (c *ControlNode) NodeType() string { return "Control" }
func (c *ControlNode) GetLine() int     { return c.Line }

type AssignStatement struct {
	Line  int
	Name  string
	Value string
}

func (a *AssignStatement) NodeType() string { return "Assign" }
func (a *AssignStatement) GetLine() int     { return a.Line }

type OutputCall struct {
	Line  int
	Relay string
	State string
}

func (o *OutputCall) NodeType() string { return "OutputCall" }
func (o *OutputCall) GetLine() int     { return o.Line }

type IfStatement struct {
	Line     int
	Left     string
	Operator string
	Right    string
	Then     Node
}

func (i *IfStatement) NodeType() string { return "IfStatement" }
func (i *IfStatement) GetLine() int     { return i.Line }

type IfElseStatement struct {
	Line     int
	LeftVar  string
	LeftOp   string
	LeftVal  string
	RightVar string
	RightOp  string
	RightVal string
	Then     Node
	Else     Node
}

func (i *IfElseStatement) NodeType() string { return "IfElseStatement" }
func (i *IfElseStatement) GetLine() int     { return i.Line }

type IfOrStatement struct {
	Line     int
	LeftVar  string
	LeftOp   string
	LeftVal  string
	RightVar string
	RightOp  string
	RightVal string
	Then     Node
}

func (i *IfOrStatement) NodeType() string { return "IfOrStatement" }
func (i *IfOrStatement) GetLine() int     { return i.Line }

type IfOrElseStatement struct {
	Line     int
	LeftVar  string
	LeftOp   string
	LeftVal  string
	RightVar string
	RightOp  string
	RightVal string
	Then     Node
	Else     Node
}

func (i *IfOrElseStatement) NodeType() string { return "IfOrElseStatement" }
func (i *IfOrElseStatement) GetLine() int     { return i.Line }

type ConstNode struct {
	Line      int
	Constants []ConstDecl
}

func (c *ConstNode) NodeType() string { return "Const" }
func (c *ConstNode) GetLine() int     { return c.Line }

type ConstDecl struct {
	Line  int
	Name  string
	Value string
}

func (c *ConstDecl) NodeType() string { return "ConstDecl" }
func (c *ConstDecl) GetLine() int     { return c.Line }
