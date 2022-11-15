package misc

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestColum(t *testing.T) {
	s := `29abd4c6e599   avalido/mpc-controller                                 "/app/mpc-controller…"   20 hours ago   Exited (0) 3 minutes ago                                                mpc-7_controller_1
cbf9a4610131   avalido/mpc-controller                                 "/app/mpc-controller…"   20 hours ago   Exited (0) 3 minutes ago                                                mpc-6_controller_1
e01f9247f5a4   avalido/mpc-controller                                 "/app/mpc-controller…"   20 hours ago   Exited (0) 3 minutes ago                                                mpc-5_controller_1
afa3c87db82a   avalido/mpc-controller                                 "/app/mpc-controller…"   20 hours ago   Exited (0) 3 minutes ago                                                mpc-4_controller_1
7ac33878b1db   avalido/mpc-controller                                 "/app/mpc-controller…"   20 hours ago   Exited (0) 3 minutes ago                                                mpc-3_controller_1
caaa9197fb13   avalido/mpc-controller                                 "/app/mpc-controller…"   20 hours ago   Exited (0) 3 minutes ago                                                mpc-2_controller_1
b68d8590dc60   avalido/mpc-controller                                 "/app/mpc-controller…"   20 hours ago   Exited (0) 3 minutes ago                                                mpc-1_controller_1
b2b61df5934f   avalido/oracle                                         "/app/oracle --rpc-u…"   20 hours ago   Exited (0) 20 hours ago                                                 mpc-1_oracle_1
7b5fa537f004   avalido/mpc-server                                     "/app/mpc-server -s …"   20 hours ago   Created                     0.0.0.0:8001->8001/tcp, :::8001->8001/tcp   mpc-1_server_1
969cc1c87db3   avalido/mpc-controller                                 "/app/mpc-controller…"   20 hours ago   Exited (0) 20 hours ago                                                 mpc_controller2
1bd9b2c789f7   avalido/mpc-controller                                 "/app/mpc-controller…"   20 hours ago   Exited (0) 20 hours ago                                                 mpc_controller1
7f0aefac00fc   avalido/mpc-controller                                 "/app/mpc-controller…"   20 hours ago   Exited (0) 20 hours ago                                                 mpc_controller7
b1ccbac8a8b7   avalido/mpc-controller                                 "/app/mpc-controller…"   20 hours ago   Exited (0) 20 hours ago                                                 mpc_controller3
056401c90dbe   avalido/mpc-controller                                 "/app/mpc-controller…"   20 hours ago   Exited (0) 20 hours ago                                                 mpc_controller4
38ce8224e957   avalido/mpc-controller                                 "/app/mpc-controller…"   20 hours ago   Exited (0) 20 hours ago                                                 mpc_controller6
e54b47f0c6fe   avalido/mpc-controller                                 "/app/mpc-controller…"   20 hours ago   Exited (0) 20 hours ago                                                 mpc_controller5
`
	expected := "29abd4c6e599 cbf9a4610131 e01f9247f5a4 afa3c87db82a 7ac33878b1db caaa9197fb13 b68d8590dc60 b2b61df5934f 7b5fa537f004 969cc1c87db3 1bd9b2c789f7 7f0aefac00fc b1ccbac8a8b7 056401c90dbe 38ce8224e957 e54b47f0c6fe"
	col := ExtractColumIntoString(s, " ", 0)
	require.Equal(t, col, expected)
}
