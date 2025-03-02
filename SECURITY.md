# Security Policy

## Reporting Vulnerabilities
If you discover a vulnerability, please report it immediately to [kmosc@protonmail.com](mailto:kmosc@protonmail.com) with all relevant details.

## Security Best Practices
- **Container Security:** Use minimal base images and run as non-root.
- **Static Analysis:** Regularly run tools like GoSec.
- **CI/CD:** Integrate vulnerability scanners such as Trivy.
- **Kubernetes Hardening:** Enforce RBAC and use network policies.
- **Communication Security:** Enforce TLS/mTLS for all external and internal communications.

## Incident Response
In the event of a security breach, our incident response plan includes:
- Immediate isolation of affected systems.
- Detailed logging and forensic analysis.
- Transparent communication with stakeholders.
