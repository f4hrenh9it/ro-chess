// Code generated by "stringer -type=comboType"; DO NOT EDIT.

package internal

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[attack3Combo-0]
	_ = x[defense3Combo-1]
}

const _comboType_name = "attack3Combodefense3Combo"

var _comboType_index = [...]uint8{0, 12, 25}

func (i comboType) String() string {
	if i < 0 || i >= comboType(len(_comboType_index)-1) {
		return "comboType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _comboType_name[_comboType_index[i]:_comboType_index[i+1]]
}