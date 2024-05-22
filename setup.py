import os
import sys
import argparse
import subprocess
import platform
import random
import tempfile
from enum import Enum

DEFAULT_HOME = os.path.expanduser("~/.cored")
DEFAULT_MONIKER = "coreum"

NETWORK_CHOICES = ['coreum-mainnet-1', 'coreum-testnet-1']
INSTALL_CHOICES = ['node', 'client']
PRUNING_CHOICES = ['default', 'nothing', 'everything']

MAINNET_VERSION = "3.0.3"
TESTNET_VERSION = "3.0.3"

# CLI arguments
parser = argparse.ArgumentParser(description="Coreum Installer")

parser.add_argument(
    "--home",
    type=str,
    help=f"Cored installation location",
)

parser.add_argument(
    '-m',
    "--moniker",
    type=str,
    help=f"Moniker name for the node (Default: '{DEFAULT_MONIKER}')",
)

parser.add_argument(
    '-v',
    '--verbose',
    action='store_true',
    help="Enable verbose output",
    dest="verbose"
)

parser.add_argument(
    '-o',
    '--overwrite',
    action='store_true',
    help="Overwrite existing Coreum home without prompt",
    dest="overwrite"
)

parser.add_argument(
    '-n',
    '--network',
    type=str,
    choices=NETWORK_CHOICES,
    help=f"Network to join: {NETWORK_CHOICES})",
)

parser.add_argument(
    '-p',
    '--pruning',
    type=str,
    choices=PRUNING_CHOICES,
    help=f"Pruning settings: {PRUNING_CHOICES})",
)

parser.add_argument(
    '-i',
    '--install',
    type=str,
    choices=INSTALL_CHOICES,
    help=f"Which installation to do: {INSTALL_CHOICES})",
)

parser.add_argument(
    "--binary_path",
    type=str,
    help=f"Path where to download the binary",
    default="/usr/local/bin"
)

parser.add_argument(
    '-c',
    '--cosmovisor',
    action='store_true',
    help="Install cosmovisor"
)

parser.add_argument(
    '-s',
    '--service',
    action='store_true',
    help="Setup systemd service (Linux only)"
)

args = parser.parse_args()

# Choices
class InstallChoice(str, Enum):
    NODE = "1"
    CLIENT = "2"

class NetworkChoice(str, Enum):
    MAINNET = "1"
    TESTNET = "2"

class PruningChoice(str, Enum):
    DEFAULT = "1"
    NOTHING = "2"
    EVERYTHING = "3"

class Answer(str, Enum):
    YES = "1"
    NO = "2"

# Network configurations
class Network:
    def __init__(self, chain_id, version, binary_url):
        self.chain_id = chain_id
        self.version = version
        self.binary_url = binary_url

TESTNET = Network(
    chain_id = "coreum-testnet-1",
    version = f"v${TESTNET_VERSION}",
    binary_url = {
        "linux": {
            "amd64": f"https://github.com/CoreumFoundation/coreum/releases/download/v{TESTNET_VERSION}/cored-linux-amd64",
            "arm64": f"https://github.com/CoreumFoundation/coreum/releases/download/v{TESTNET_VERSION}/cored-linux-arm64"
        },
    },
)

MAINNET = Network(
    chain_id = "coreum-mainnet-1",
    version = f"v{MAINNET_VERSION}",
    binary_url = {
        "linux": {
            "amd64": f"https://github.com/CoreumFoundation/coreum/releases/download/v{MAINNET_VERSION}/cored-linux-amd64",
            "arm64": f"https://github.com/CoreumFoundation/coreum/releases/download/v{MAINNET_VERSION}/cored-linux-arm64"
        },
    },
)

COSMOVISOR_URL = {
    "linux": {
        "amd64": "https://github.com/cosmos/cosmos-sdk/releases/download/cosmovisor%2Fv1.5.0/cosmovisor-v1.5.0-linux-amd64.tar.gz",
        "arm64": "https://github.com/cosmos/cosmos-sdk/releases/download/cosmovisor%2Fv1.5.0/cosmovisor-v1.5.0-linux-arm64.tar.gz"
    }
}
# Terminal utils

class bcolors:
    OKGREEN = '\033[92m'
    RED = '\033[91m'
    ENDC = '\033[0m'
    PURPLE = '\033[95m'

def clear_screen():
    os.system('clear')

# Messages

def welcome_message():
    print(bcolors.OKGREEN + """
 Welcome to the Coreum node installer!


For more information, please visit https://docs.coreum.dev

If you have an old Coreum installation, 
- backup any important data before proceeding
- ensure that no cored services are running in the background
""" + bcolors.ENDC)


def client_complete_message(home):
    print(bcolors.OKGREEN + """
✨ Congratulations! You have successfully completed setting up Coreum client! ✨
""" + bcolors.ENDC)

    print("🧪 Try running: " + bcolors.OKGREEN + f"cored status --home {home}" + bcolors.ENDC)
    print()


def node_complete_message(using_cosmovisor, using_service, home):
    print(bcolors.OKGREEN + """
✨ Congratulations! You have successfully completed setting up Coreum node! ✨
""" + bcolors.ENDC)
    
    if using_service:

        if using_cosmovisor:
            print("🧪 To start the cosmovisor service run: ")
            print(bcolors.OKGREEN + f"sudo systemctl start cosmovisor" + bcolors.ENDC)
        else:
            print("🧪 To start the cored service run: ")
            print(bcolors.OKGREEN + f"sudo systemctl start cored" + bcolors.ENDC)

    else:
        if using_cosmovisor:
            print("🧪 To start cosmovisor run: ")
            print(bcolors.OKGREEN + f"DAEMON_NAME=cored DAEMON_HOME={home} cosmovisor run start" + bcolors.ENDC)
        else:
            print("🧪 To start cored run: ")
            print(bcolors.OKGREEN + f"cored start --home {home}" + bcolors.ENDC)


    
    print()

# Options

def select_install():

    # Check if setup is specified in args
    if args.install:
        if args.install == "node":
            choice = InstallChoice.NODE
        elif args.install == "client":
            choice = InstallChoice.CLIENT
        else:
            print(bcolors.RED + f"Invalid setup {args.install}. Please choose a valid setup.\n" + bcolors.ENDC)
            sys.exit(1)
    
    else:

        print(bcolors.OKGREEN + """
Please choose the desired installation:

    1) node         - run Coreum node and join mainnet or testnet
    2) client       - setup cored to query a public node

💡 You can select the installation using the --install flag.
        """ + bcolors.ENDC)

        while True:
            choice = input("Enter your choice, or 'exit' to quit: ").strip()

            if choice.lower() == "exit":
                print("Exiting the program...")
                sys.exit(0)

            if choice not in [InstallChoice.NODE, InstallChoice.CLIENT]:
                print("Invalid input. Please choose a valid option.")
            else:
                break
            
        if args.verbose:
            clear_screen()
            print(f"Chosen install: {INSTALL_CHOICES[int(choice) - 1]}")

    clear_screen()
    return choice


def select_network():
    """
    Selects a network based on user input or command-line arguments.

    Returns:
        chosen_network (Network): The chosen network, either MAINNET or TESTNET.

    Raises:
        SystemExit: If an invalid network is specified or the user chooses to exit the program.
    """

    # Check if network is specified in args
    if args.network:
        if args.network == MAINNET.chain_id:
            choice = NetworkChoice.MAINNET
        elif args.network == TESTNET.chain_id:
            choice = NetworkChoice.TESTNET
        else:
            print(bcolors.RED + f"Invalid network {args.network}. Please choose a valid network." + bcolors.ENDC)
            sys.exit(1)

    # If not, ask the user to choose a network
    else:
        print(bcolors.OKGREEN + f"""
Please choose the desired network:

    1) Mainnet ({MAINNET.chain_id})
    2) Testnet ({TESTNET.chain_id})

💡 You can select the network using the --network flag.
""" + bcolors.ENDC)

        while True:
            choice = input("Enter your choice, or 'exit' to quit: ").strip()

            if choice.lower() == "exit":
                print("Exiting the program...")
                sys.exit(0)

            if choice not in [NetworkChoice.MAINNET, NetworkChoice.TESTNET]:
                print(bcolors.RED + "Invalid input. Please choose a valid option. Accepted values: [ 1 , 2 ] \n" + bcolors.ENDC)
            else:
                break
        
    if args.verbose:
        clear_screen()
        print(f"Chosen network: {NETWORK_CHOICES[int(choice) - 1]}")

    clear_screen()

    if choice == NetworkChoice.TESTNET:
        return TESTNET

    return MAINNET


def select_home():
    """
    Selects the path for running the 'cored init --home <SELECTED_HOME>' command.

    Returns:
        home (str): The selected path.

    """
    if args.home:
        home = args.home
    else:
        default_home = os.path.expanduser("~/.cored")
        print(bcolors.OKGREEN + f"""
Do you want to install Coreum in the default location?:

    1) Yes, use default location {DEFAULT_HOME} (recommended)
    2) No, specify custom location

💡 You can specify the home using the --home flag.
""" + bcolors.ENDC)

        while True:
            choice = input("Enter your choice, or 'exit' to quit: ").strip()

            if choice.lower() == "exit":
                print("Exiting the program...")
                sys.exit(0)

            if choice == Answer.YES:
                home = default_home
                break

            elif choice == Answer.NO:
                while True:
                    custom_home = input("Enter the path for Coreum home: ").strip()
                    if custom_home != "":
                        home = custom_home
                        break
                    else:
                        print("Invalid path. Please enter a valid directory.")
                break
            else:
                print("Invalid choice. Please enter 1 or 2.")

    clear_screen()
    return home


def select_moniker():
    """
    Selects the moniker for the Coreum node.

    Returns:
        moniker (str): The selected moniker.

    """
    if args.moniker:
        moniker = args.moniker
    else:
        print(bcolors.OKGREEN + f"""
Do you want to use the default moniker?

    1) Yes, use default moniker ({DEFAULT_MONIKER})
    2) No, specify custom moniker

💡 You can specify the moniker using the --moniker flag.
""" + bcolors.ENDC)

        while True:
            choice = input("Enter your choice, or 'exit' to quit: ").strip()

            if choice.lower() == "exit":
                print("Exiting the program...")
                sys.exit(0)

            if choice == Answer.YES:
                moniker = DEFAULT_MONIKER
                break
            elif choice == Answer.NO:
                while True:
                    custom_moniker = input("Enter the custom moniker: ")
                    if custom_moniker.strip() != "":
                        moniker = custom_moniker
                        break
                    else:
                        print("Invalid moniker. Please enter a valid moniker.")
                break
            else:
                print("Invalid choice. Please enter 1 or 2.")

    clear_screen()
    return moniker


def initialize_home(network, home, moniker):
    """
    Initializes the Coreum home directory with the specified moniker.

    Args:
        network (Network): Selected network.
        home (str): The chosen home directory.
        moniker (str): The moniker for the Coreum node.

    """
    if not args.overwrite:

        while True:
            print(bcolors.OKGREEN + f"""
Do you want to initialize the Coreum home directory at '{home}'?
            """ + bcolors.ENDC, end="")

            print(bcolors.RED + f"""
⚠️ All contents of the directory will be deleted.
            """ + bcolors.ENDC, end="")

            print(bcolors.OKGREEN + f"""
    1) Yes, proceed with initialization
    2) No, quit

💡 You can overwrite the Coreum home using --overwrite flag.
            """ + bcolors.ENDC)
            
            choice = input("Enter your choice, or 'exit' to quit: ").strip()

            if choice.lower() == "exit":
                print("Exiting the program...")
                sys.exit(0)

            if choice == Answer.YES:
                break

            elif choice == Answer.NO:
                sys.exit(0)

            else:
                print("Invalid choice. Please enter 1 or 2.")
    
    print(f"Initializing Coreum home directory at '{home}'...")
    try:
        subprocess.run(
            ["rm", "-rf", home],
            stderr=subprocess.DEVNULL, check=True)
        
        subprocess.run(
            ["cored", "init", moniker,  "-o", "--home", home, "--chain-id", network.chain_id],
            stderr=subprocess.DEVNULL, check=True)

        print("Initialization completed successfully.")

    except subprocess.CalledProcessError as e:
        print("Initialization failed.")
        print("Please check if the home directory is valid and has write permissions.")
        print(e)
        sys.exit(1)

    clear_screen()


def select_pruning(network, home):
    """
    Allows the user to choose pruning settings and performs actions based on the selected option.

    """

    # Check if pruning settings are specified in args
    if args.pruning:
        if args.pruning == "default":
            choice = PruningChoice.DEFAULT
        elif args.pruning == "nothing":
            choice = PruningChoice.NOTHING
        elif args.pruning ==  "everything":
            choice = PruningChoice.EVERYTHING
        else:
            print(bcolors.RED + f"Invalid pruning setting {args.pruning}. Please choose a valid setting.\n" + bcolors.ENDC)
            sys.exit(1)
    
    else:

        print(bcolors.OKGREEN + """
Please choose your desired pruning settings:

    1) Default: (keep last 100,000 states to query the last week worth of data and prune at 100 block intervals)
    2) Nothing: (keep everything, select this if running an archive node)
    3) Everything: (keep last 10,000 states and prune at a random prime block interval)

💡 You can select the pruning settings using the --pruning flag.
    """ + bcolors.ENDC)

        while True:
            choice = input("Enter your choice, or 'exit' to quit: ").strip()

            if choice.lower() == "exit":
                print("Exiting the program...")
                sys.exit(0)

            if choice not in [PruningChoice.DEFAULT, PruningChoice.NOTHING, PruningChoice.EVERYTHING]:
                print("Invalid input. Please choose a valid option.")
            else:
                break
            
        if args.verbose:
            clear_screen()
            print(f"Chosen setting: {PRUNING_CHOICES[int(choice) - 1]}")
    
    app_toml = os.path.join(home,network.chain_id, "config", "app.toml")

    if choice == PruningChoice.DEFAULT:
        # Nothing to do
        pass

    elif choice == PruningChoice.NOTHING:
        subprocess.run(["sed -i -E 's/pruning = \"default\"/pruning = \"nothing\"/g' " + app_toml], shell=True)

    elif choice == PruningChoice.EVERYTHING:
        primeNum = random.choice([x for x in range(11, 97) if not [t for t in range(2, x) if not x % t]])
        subprocess.run(["sed -i -E 's/pruning = \"default\"/pruning = \"custom\"/g' " + app_toml], shell=True)
        subprocess.run(["sed -i -E 's/pruning-keep-recent = \"0\"/pruning-keep-recent = \"10000\"/g' " + app_toml], shell=True)
        subprocess.run(["sed -i -E 's/pruning-interval = \"0\"/pruning-interval = \"" + str(primeNum) + "\"/g' " + app_toml], shell=True)
    
    else:
        print(bcolors.RED + f"Invalid pruning setting {choice}. Please choose a valid setting.\n" + bcolors.ENDC)
        sys.exit(1)

    clear_screen()


def download_binary(network):
    """
    Downloads the binary for the specified network based on the operating system and architecture.

    Args:
        network (Network): The network type, either MAINNET or TESTNET.

    Raises:
        SystemExit: If the binary download URL is not available for the current operating system and architecture.

    """
    operating_system = platform.system().lower()
    architecture = platform.machine()

    if architecture == "x86_64":
        architecture = "amd64"
    elif architecture == "aarch64":
        architecture = "arm64"

    if architecture not in ["arm64", "amd64"]:
        print(f"Unsupported architecture {architecture}.")
        sys.exit(1)

    binary_urls = network.binary_url
    if operating_system in binary_urls and architecture in binary_urls[operating_system]:
        binary_url = binary_urls[operating_system][architecture]
    else:
        print(f"Binary download URL not available for {operating_system}/{architecture}")
        sys.exit(0)

    try:   
        binary_path = os.path.join(args.binary_path, "cored")

        print("Downloading " + bcolors.PURPLE+ "cored" + bcolors.ENDC, end="\n\n")
        print("from " + bcolors.OKGREEN + f"{binary_url}" + bcolors.ENDC, end=" ")
        print("to " + bcolors.OKGREEN + f"{binary_path}" + bcolors.ENDC)
        print()
        print(bcolors.OKGREEN + "💡 You can change the path using --binary_path" + bcolors.ENDC)

        subprocess.run(["wget", binary_url,"-q", "-O", "/tmp/cored"], check=True)
        os.chmod("/tmp/cored", 0o755)

        if platform.system() == "Linux":
            subprocess.run(["sudo", "mv", "/tmp/cored", binary_path], check=True)
            subprocess.run(["sudo", "chown", f"{os.environ['USER']}:{os.environ['USER']}", binary_path], check=True)
            subprocess.run(["sudo", "chmod", "+x", binary_path], check=True)
        else:
            subprocess.run(["mv", "/tmp/cored", binary_path], check=True)

        # Test binary 
        subprocess.run(["cored", "version"], check=True)

        print("Binary downloaded successfully.")

    except subprocess.CalledProcessError as e:
        print(e)
        print("Failed to download the binary.")
        sys.exit(1)

    clear_screen()


def download_cosmovisor(network, home):
    """
    Downloads and installs cosmovisor.

    Returns:
        use_cosmovisor(bool): Whether to use cosmovisor or not.

    """
    if not args.cosmovisor:
        print(bcolors.OKGREEN + f"""
Do you want to install cosmovisor?

    1) Yes, download and install cosmovisor (default)
    2) No

💡 You can specify the cosmovisor setup using the --cosmovisor flag.
""" + bcolors.ENDC)

        while True:
            choice = input("Enter your choice, or 'exit' to quit: ").strip()

            if choice.lower() == "exit":
                print("Exiting the program...")
                sys.exit(0)

            if choice == Answer.YES:
                break
            elif choice == Answer.NO:
                print("Skipping cosmovisor installation.")
                clear_screen()
                return False
            else:
                print("Invalid choice. Please enter 1 or 2.")

    # Download and install cosmovisor
    operating_system = platform.system().lower()
    architecture = platform.machine()

    if architecture == "x86_64":
        architecture = "amd64"
    elif architecture == "aarch64":
        architecture = "arm64"

    if architecture not in ["arm64", "amd64"]:
        print(f"Unsupported architecture {architecture}.")
        sys.exit(1)
    
    if operating_system in COSMOVISOR_URL and architecture in COSMOVISOR_URL[operating_system]:
        cosmovisor_url = COSMOVISOR_URL[operating_system][architecture]
    else:
        print(f"Binary download URL not available for {os}/{architecture}")
        sys.exit(0)

    try:   
        binary_path = os.path.join(args.binary_path, "cosmovisor")

        print("Downloading " + bcolors.PURPLE+ "cosmovisor" + bcolors.ENDC, end="\n\n")
        print("from " + bcolors.OKGREEN + f"{cosmovisor_url}" + bcolors.ENDC, end=" ")
        print("to " + bcolors.OKGREEN + f"{binary_path}" + bcolors.ENDC)
        print()
        print(bcolors.OKGREEN + "💡 You can change the path using --binary_path" + bcolors.ENDC)

        clear_screen()
        temp_dir = tempfile.mkdtemp()
        temp_tar_path = os.path.join(temp_dir, "cosmovisor.tar.gz")
        temp_binary_path = os.path.join(temp_dir, "cosmovisor")

        subprocess.run(["wget", cosmovisor_url,"-q", "-O", temp_tar_path], check=True)
        subprocess.run(["tar", "-xf", temp_tar_path, "-C", temp_dir], check=True)
        os.chmod(temp_binary_path, 0o755)

        if platform.system() == "Linux":
            subprocess.run(["sudo", "mv", temp_binary_path, binary_path], check=True)
            subprocess.run(["sudo", "chown", f"{os.environ['USER']}:{os.environ['USER']}", binary_path], check=True)
            subprocess.run(["sudo", "chmod", "+x", binary_path], check=True)
        else:
            subprocess.run(["mv", temp_binary_path, binary_path], check=True)

        # Test binary 
        subprocess.run(["cosmovisor", "help"], check=True)

        print("Binary downloaded successfully.")

    except subprocess.CalledProcessError:
        print("Failed to download the binary.")
        sys.exit(1)

    clear_screen()

    # Initialize cosmovisor
    print("Setting up cosmovisor directory...")

    # Set environment variables
    env = {
        "DAEMON_NAME": "cored",
        "DAEMON_HOME": os.path.join(home, network.chain_id)
    }

    try:
        subprocess.run(["/usr/local/bin/cosmovisor", "init", "/usr/local/bin/cored"], check=True, env=env)
    except subprocess.CalledProcessError:
        print("Failed to initialize cosmovisor.")
        sys.exit(1)

    clear_screen()
    return True


def setup_cosmovisor_service(network, home):
    """
    Setup cosmovisor service on Linux.
    """

    operating_system = platform.system()

    if operating_system != "Linux":
        return False
    
    if not args.service:
        print(bcolors.OKGREEN + f"""
Do you want to setup cosmovisor as a background service?

    1) Yes, setup cosmovisor as a service
    2) No

💡 You can specify the service setup using the --service flag.
""" + bcolors.ENDC)

        while True:
            choice = input("Enter your choice, or 'exit' to quit: ").strip()

            if choice.lower() == "exit":
                print("Exiting the program...")
                sys.exit(0)

            if choice == Answer.YES:
                break
            elif choice == Answer.NO:
                return
    
    user = os.environ.get("USER")
    cosmovisor_home = os.path.join(home, network.chain_id)
    
    unit_file_contents = f"""[Unit]
Description=Cosmovisor daemon
After=network-online.target

[Service]
Environment="DAEMON_NAME=cored"
Environment="DAEMON_HOME={cosmovisor_home}"
Environment="DAEMON_RESTART_AFTER_UPGRADE=true"
Environment="DAEMON_ALLOW_DOWNLOAD_BINARIES=false"
Environment="DAEMON_LOG_BUFFER_SIZE=512"
Environment="UNSAFE_SKIP_BACKUP=true"
User={user}
ExecStart=/usr/local/bin/cosmovisor run start --home {home}
Restart=always
RestartSec=3
LimitNOFILE=infinity
LimitNPROC=infinity

[Install]
WantedBy=multi-user.target
"""

    unit_file_path = "/lib/systemd/system/cosmovisor.service"

    with open("cosmovisor.service", "w") as f:
        f.write(unit_file_contents)

    subprocess.run(["sudo", "mv", "cosmovisor.service", unit_file_path])
    subprocess.run(["sudo", "systemctl", "daemon-reload"])
    subprocess.run(["systemctl", "restart", "systemd-journald"])

    clear_screen()
    return True


def setup_cored_service(home):
    """
    Setup cored service on Linux.
    """

    operating_system = platform.system()

    if operating_system != "Linux":
        return False

    if not args.service:
        print(bcolors.OKGREEN + """
Do you want to set up cored as a background service?

    1) Yes, set up cored as a service
    2) No

💡 You can specify the service setup using the --service flag.
""" + bcolors.ENDC)

        while True:
            choice = input("Enter your choice, or 'exit' to quit: ").strip()

            if choice.lower() == "exit":
                print("Exiting the program...")
                sys.exit(0)

            if choice == Answer.YES:
                break
            elif choice == Answer.NO:
                return
    
    user = os.environ.get("USER")
    
    unit_file_contents = f"""[Unit]
Description=Coreum Daemon
After=network-online.target

[Service]
User={user}
ExecStart=/usr/local/bin/cored start --home {home}
Restart=always
RestartSec=3
LimitNOFILE=infinity
LimitNPROC=infinity

[Install]
WantedBy=multi-user.target
"""

    unit_file_path = "/lib/systemd/system/cored.service"

    with open("cored.service", "w") as f:
        f.write(unit_file_contents)

    subprocess.run(["sudo", "mv", "cored.service", unit_file_path])
    subprocess.run(["sudo", "systemctl", "daemon-reload"])
    subprocess.run(["systemctl", "restart", "systemd-journald"])

    clear_screen()
    return True


def main():

    welcome_message()

    # Start the installation
    chosen_install = select_install()

    if chosen_install == InstallChoice.NODE:
        network = select_network()
        download_binary(network)
        home = select_home()
        moniker = select_moniker()
        initialize_home(network, home, moniker)
        using_cosmovisor = download_cosmovisor(network, home)
        select_pruning(network, home)
        if using_cosmovisor:
            using_service = setup_cosmovisor_service(network, home)
        else:
            using_service = setup_cored_service(home)
        node_complete_message(using_cosmovisor, using_service, home)

    elif chosen_install == InstallChoice.CLIENT:
        network = select_network()
        download_binary(network)
        home = select_home()
        moniker = select_moniker()
        initialize_home(network, home, moniker)
        client_complete_message(home)

main()
