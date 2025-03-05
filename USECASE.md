# **Thunder - Use Cases**

## **Overview**
Thunder is a minimalist backend framework in Go, designed for **gRPC-Gateway-powered** applications with **Prisma ORM** and **Kubernetes**. It is built for **scalable microservices** and **high-performance API development**.

---

## **When to Use Thunder?**

### **1. High-Performance API Development**
- If you need a **gRPC-first API** with a **RESTful interface** via **gRPC-Gateway**.
- When performance and **low latency** are critical (gRPC offers faster communication than REST).
- When you want to ensure **strongly-typed** APIs with **protobufs**.

### **2. Microservices Architecture**
- When you are developing a **microservices-based backend** where services need to communicate efficiently.
- If you need **inter-service communication via gRPC**.
- When deploying to **Kubernetes** with built-in **service discovery and scaling**.

### **3. Database Management with Prisma**
- If you want to use **Prisma ORM** with Go for **type-safe queries and database migrations**.
- When you need an **easy-to-manage database schema**.
- If you want to support **multiple databases** (PostgreSQL, MySQL, SQLite, etc.).

### **4. Lightweight Backend Alternative**
- If you are looking for a minimalist yet **powerful alternative** to traditional Go web frameworks like **Gin or Echo**.
- When you want a **fast, simple, and modular** backend without unnecessary overhead.

### **5. Kubernetes & Cloud-Native Applications**
- If you are deploying services in **Kubernetes** and need a **scalable backend**.
- When working with **containerized environments using Docker**.
- If you require **automatic service scaling and load balancing**.

---

## **Example Use Cases**

| Use Case | How Thunder Helps |
|----------|------------------|
| **Building gRPC & REST APIs** | Thunder supports **gRPC + REST API generation** with **gRPC-Gateway**. |
| **Microservices Communication** | Uses **gRPC** for **fast, efficient, and scalable inter-service communication**. |
| **Database-Driven Applications** | Integrated with **Prisma ORM** for **database management and migrations**. |
| **Cloud & Kubernetes Deployments** | Designed for **Kubernetes**, supports **containerized deployments**. |
| **Backend for IoT & Realtime Applications** | Supports **bidirectional streaming**, making it ideal for **real-time systems**. |

---

## **When Not to Use Thunder?**
- If you need a **traditional REST-only API** without gRPC (use **Gin, Fiber, or Echo** instead).
- If you want a **feature-heavy web framework** with built-in middleware.
- If you are not deploying on **Kubernetes** or need a monolithic backend.

---

## **Conclusion**
**Thunder** is the perfect choice for **modern, scalable, and high-performance backend services**. Whether you're working with **microservices, real-time applications, or cloud-native deployments**, Thunder provides a **minimalist yet powerful** solution. 🚀
