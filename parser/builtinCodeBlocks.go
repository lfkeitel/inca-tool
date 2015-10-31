package parser

var builtinBlocks = map[string]string{
	// Special builtin that basically stops compiling, used for testing
	"nil": "",

	"juniper-configure": `
expect {
    "*assword:" { send_error "$hostname Authentication failed\n"; exit 1 }
    "0%" {
        send "cli\n"
        expect ">"
    }
    ">"
}
send "configure\n"
expect "#"
`,

	// Exit without commiting changes
	"juniper-exit-nocommit": `
send "exit\n"
expect {
    "Exit with uncommitted changes?" { send "yes\n"; expect ">" }
    ">"
}
send "exit\n"
expect {
	"0%" { send "exit\n" }
	eof {}
}
`,

	// Attempt a commit, if failure rollback and alert user
	"juniper-commit-rollback-failed": `
set oldTimeout $timeout
set timeout 30
send "commit\n"
expect {
	-re "error|failed" {
		send "rollback\n"
		expect "*#"
		send "exit\n"
		expect "*>"
		send "exit\n"
		send_error "$hostname failed to commit changes"
		exit 1
	}
	"commit complete"
}
set timeout $oldTimeout
`,

	"cisco-enable-mode": `
expect {
	default { send_error "$hostname Enable Mode Failed - Check Password\n"; exit 1 }
	"#" {}
	">" {
		send "enable\n"
		expect "*assword"
		send "$enablepassword\n"
		expect {
			"% Access denied" {
				send_error "$hostname Enable Mode Failed - Check Password\n"
				exit 1
			}
			"#"
		}
	}
}
`,

	// Exit configure mode, wr mem, then continue
	"cisco-end-wrmem": `
send "end\n"
expect "#"
set oldTimeout $timeout
set timeout 30
send "wr mem\n"
expect {
	default {
		send "exit\n"
		send_error "$hostname failed to save configuration"
		exit 1
	}
	"[OK]"
}
set timeout $oldTimeout
`,
}
