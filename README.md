# GitID - Git Identity Manager

GitID is a lightweight CLI tool that manages Git's `includeIf` configuration feature, making it easy to automatically use different Git identities based on directory location.

## Quick Start

```bash
# 1. Build GitID (or install from homebrew: brew tap Kir-Push/gitid; brew install gitid)
go build -o gitid cmd/gitid/main.go

# 2. Add your identities
./gitid add work --name "John Doe" --email john@company.com --path ~/work
./gitid add personal --name "johndoe" --email personal@gmail.com --path ~/personal

# 4. Test it works
cd ~/work/some-project
git config user.email  # Shows: john@company.com

cd ~/personal/my-blog  
git config user.email  # Shows: personal@gmail.com
```


## 📦 Installation

### From Source (Current Method)

```bash
# Add tap
brew tap Kir-Push/gitid

# Install
brew install gitid
```


## 🔧 Commands

### Core Commands

#### `gitid add`
Add a new Git identity with directory path mapping.

```bash
gitid add [identity-name] --name "Display Name" --email "email@domain.com" --path "/path/to/directory"

# Examples
gitid add work --name "John Doe" --email "john@company.com" --path "~/work"
gitid add personal --name "johndoe" --email "me@gmail.com" --path "~/personal"
gitid add opensource --name "johndoe" --email "johndoe@users.noreply.github.com" --path "~/opensource"
```

**Flags:**
- `--name, -n`: Git user name (required)
- `--email, -e`: Git user email (required) 
- `--path, -p`: Directory path for this identity (required)

#### `gitid list`
Display all configured Git identities and their settings.

```bash
gitid list
```

Example output:
```
📋 Configured identities:
  work
    Name: John Doe
    Email: john@company.com
    Paths: ~/work

  personal
    Name: johndoe
    Email: me@gmail.com
    Paths: ~/personal
```

#### `gitid status`
Show which identity is currently active in the current directory.

```bash
gitid status
```

Example output:
```
📍 Current directory: /Users/john/work/project
✅ Matching identities:
  - work (john@company.com)
```

#### `gitid test`
Test which Git identity would be applied for a specific directory path.

```bash
gitid test [path]

# Examples
gitid test ~/work/new-project
gitid test ~/personal/blog
gitid test ~/random/path
```

#### `gitid remove`
Remove a Git identity and clean up its configuration.

```bash
gitid remove [identity-name]

# Example
gitid remove work
```