package exitcodes

type ExitCode int

// See: https://tldp.org/LDP/abs/html/exitcodes.html
// tl;dr 0-2, 126-165, and 255+ are reserved
const (
	// Success exit code for all successful outcomes, any exit code other than this 0 means failure.
	Success ExitCode = 0
	// Error exit code for general errors that do not need to be differentiable by the user.
	Error ExitCode = 1
	// Conflict exit code for when a command fails due to a conflict, ex: a deployment is already in progress
	Conflict ExitCode = 3

	// The following exit codes are reserved and not to be used by the CLI
	_ ExitCode = 2
	_ ExitCode = 126
	_ ExitCode = 127
	_ ExitCode = 128
	_ ExitCode = 129
	_ ExitCode = 130
	_ ExitCode = 131
	_ ExitCode = 132
	_ ExitCode = 133
	_ ExitCode = 134
	_ ExitCode = 135
	_ ExitCode = 136
	_ ExitCode = 137
	_ ExitCode = 138
	_ ExitCode = 139
	_ ExitCode = 140
	_ ExitCode = 141
	_ ExitCode = 142
	_ ExitCode = 143
	_ ExitCode = 144
	_ ExitCode = 145
	_ ExitCode = 146
	_ ExitCode = 147
	_ ExitCode = 148
	_ ExitCode = 149
	_ ExitCode = 150
	_ ExitCode = 151
	_ ExitCode = 152
	_ ExitCode = 153
	_ ExitCode = 154
	_ ExitCode = 155
	_ ExitCode = 156
	_ ExitCode = 157
	_ ExitCode = 158
	_ ExitCode = 159
	_ ExitCode = 160
	_ ExitCode = 161
	_ ExitCode = 162
	_ ExitCode = 163
	_ ExitCode = 164
	_ ExitCode = 165
	_ ExitCode = 255
)
