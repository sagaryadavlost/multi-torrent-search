# Contributing to TorrentSearch Web

First off, thanks for taking the time to contribute! ❤️

All types of contributions are encouraged and valued. If you do know how to code and want to contribute see below and if you like the project but don't know how to code, you can support the project in other easy ways:

- Star the project
- Help to translate to other languages
- Report bugs
- Suggest new features or enhancements

## Setting up the project

### Prerequisites

- **Go**: Install Go 1.21 or later. You can download it from [here](https://go.dev/dl/).
- **Node.js**: Install Node.js 18 or later. You can download it from [here](https://nodejs.org/).
- **Git**: Install Git to clone and manage the repository.

### Steps

1. **Fork the repository**
   - Click the `Fork` button on the top right of the repository page to create your own copy.
2. **Clone the forked repository**
   ```shell
   git clone https://github.com/<your-username>/multi-torrent-search
   cd multi-torrent-search
   ```
3. **Build the frontend**
   ```shell
   cd web
   npm install
   npm run build
   cd ..
   ```
4. **Build the backend**
   ```shell
   go build -o torrentsearch-web ./cmd/server/
   ```
5. **Run the app**
   ```shell
   ./torrentsearch-web
   ```
   - Open your browser and navigate to `http://localhost:8080`

### Development guidelines

#### Code styles

**Backend (Go):**
- Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` for formatting
- Use meaningful names for variables, functions, and types
- Group related imports and remove unused ones

**Frontend (React/JS):**
- Follow [Airbnb JavaScript Style Guide](https://github.com/airbnb/javascript)
- Use ESLint and Prettier for formatting
- Use meaningful names for variables, functions, components
- Use functional components with hooks

### Commit message style

Use [Conventional Commits](https://www.conventionalcommits.org/) format as much as possible:

```
<type>: <description>
```

Common types: `feat`, `fix`, `refactor`, `chore`, `docs`, `test`.

### Adding new torrent providers

1. Create a new file in `internal/providers/` (e.g., `newprovider.go`)
2. Implement the `Provider` interface from `internal/providers/interface.go`
3. Register the provider in `internal/providers/registry.go`
4. Add provider config to `configs/config.yaml` if needed

### Creating pull request

- Once you commit your changes, push your branch to your forked repository.
- Open your forked repository and click `Contribute > Open pull request`.
- Provide a clear title and description for your pull request.
- Reference any related issues (e.g., "Fixes #123").

Thank you for contributing to TorrentSearch Web! 🚀