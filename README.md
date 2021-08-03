# GoNetic binary
This repository contains a static binary of GoNetic for 64-bit linux.
The content of this repository is only licensed for non-commercial academic use. For other use cases please contact us.

### requirements and setup
 - A 64-bit linux computer.

 - c2d compiler:
   - We are in no way affiliated with the c2d project. See the c2d manual for licensing information of the c2d compiler: "The c2d compiler is licensed only for non–commercial, research and educational use."
   - Get the linux binary `c2d_linux` here: <http://reasoning.cs.ucla.edu/c2d/> and place it in the `etc/` directory
   - Install `libc6:i386` because the c2d binary is a 32-bit executable
   - We are looking into removing this dependency for future versions of GoNetic.

### usage
`./gonetic QTL -h`

### example
`./gonetic QTL -q etc -n sample/network.txt -m sample/mutations.csv`
