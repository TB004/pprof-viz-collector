# Contributing to pprof-viz-collector

First off, thank you for considering contributing to `pprof-viz-collector`! It's people like you who make the open-source community such an amazing place to learn, inspire, and create.

## 🌟 Our Expectations

### Community & Conduct
- **Stay Friendly and Helpful**: We aim to build a welcoming community. Always be respectful, patient, and constructive in your interactions.
- **Vibe-Coding Policy**: We prioritize technical clarity and robust engineering over "vibe-coded" solutions. Every change should have a clear technical rationale.
- **AI-Assisted Contributions**: You are welcome to use AI tools (like GitHub Copilot, ChatGPT, etc.) to assist your work. However, **never trust AI blindly**. You are responsible for every line of code you submit. Every AI-generated snippet must be rigorously checked, corrected, and verified to meet our standards.

### Code Quality
- **High Standards**: We require clean, idiomatic Go code. 
- **Documentation**: All new exported functions, types, and significant internal logic must be documented.
- **Testing**: A contribution is incomplete without verification logic. Please add unit tests for new features and reproduction tests for bug fixes.

## 🛠️ Development Workflow

We use [just](https://github.com/casey/just) as our command runner. Please ensure it is installed before you begin.

1. **Fork and Clone**:
   ```bash
   git clone https://github.com/your-username/pprof-viz-collector.git
   cd pprof-viz-collector
   ```

2. **Create a Branch**:
   ```bash
   git checkout -b feature/your-feature-name
   ```

3. **Development & Verification**:
   Before submitting, ensure your changes pass all local checks:
   ```bash
   just test  # Run unit tests
   just lint  # Run golangci-lint
   just cover # Check test coverage
   just vuln  # Run vulnerability scan
   ```

4. **Commits**:
   - Keep commit messages concise and descriptive.
   - Use the imperative mood (e.g., "Add histogram support" instead of "Added histogram support").
   - Link to relevant issues if applicable.

5. **Submit a Pull Request**:
   - Fill out the PR template completely.
   - Ensure your CI build passes.
   - Be prepared for code review and feedback.

## ⚖️ License
By contributing, you agree that your contributions will be licensed under the project's [MIT License](LICENSE).
