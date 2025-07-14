# Odin DNS: Demo DNS Server in Go

![Go](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![MySQL](https://img.shields.io/badge/MySQL-005C84?style=for-the-badge&logo=mysql&logoColor=white)
![Redis](https://img.shields.io/badge/Redis-CC2921?style=for-the-badge&logo=redis&logoColor=white)
![SvelteKit](https://img.shields.io/badge/SvelteKit-FF3E00?style=for-the-badge&logo=svelte&logoColor=white)
![Caddy](https://img.shields.io/badge/Caddy-248043?style=for-the-badge&logo=caddy&logoColor=white)

Odin DNS ist ein Demonstrationsprojekt, das meine Fähigkeiten im Bereich Backend-Entwicklung mit einem Schwerpunkt auf Go-Service-Entwicklung unterstreicht.

**Wichtiger Hinweis:** Dieses Projekt ist **nicht** für den Produktionseinsatz als DNS-Server gedacht. Aus Zeitgründen wurden viele sicherheitsrelevante Aspekte nicht oder nur sehr begrenzt implementiert. Das primäre Ziel ist die Demonstration meines Verständnisses von Backend-Entwicklung und meiner Fähigkeit, auch komplexere Funktionalitäten umzusetzen.

---

### Kernfunktionen von Odin DNS

Odin DNS bietet folgende Funktionalitäten:

- **DNS-Anfragen:** Empfang und Verarbeitung von DNS-Anfragen.
- **Benutzerdefinierter Parser:** Eigens entwickelter Parser zur Zerlegung und Verarbeitung von DNS-Anfragen (keine externen DNS-Paketbibliotheken).
- **Caching-Schicht:** Schnelle Auflösung von Domains durch einen Redis-Cache.
- **Persistente Speicherung:** Sollte eine Domain nicht im Cache gefunden werden, erfolgt ein Nachschlagen in einer MySQL-Datenbank.
- **Spezifikationskonforme DNS-Antworten:** Generierung von DNS-Antworten, die den DNS-Spezifikationen entsprechen.

---

### Aktuelle Einschränkungen

Bitte beachten Sie folgende Einschränkungen von Odin DNS:

- **Kein rekursiver DNS-Server:** Odin DNS agiert nicht als rekursiver DNS-Server.
- **Begrenzte Sicherheitsfunktionen:** Das Projekt verfügt über sehr geringe bis keine Implementierung von Sicherheitsfeatures.

---

### Zusätzliche Features

Neben dem DNS-Server selbst bietet das Projekt weitere Komponenten:

- **API zur Zonenverwaltung:** Eine integrierte API ermöglicht die Verwaltung von DNS-Zonen und Einträgen. Die detaillierte API-Definition finden Sie unter:
  - [https://api.odin-demo.drinkuth.online/swagger](https://api.odin-demo.drinkuth.online/swagger) (Swagger UI)

- **Webinterface (SvelteKit):** Zur einfachen Interaktion mit der API steht ein simples Webinterface zur Verfügung:
  - [https://odin-demo.drinkuth.online](https://odin-demo.drinkuth.online)

---

### Projektkontext

Dieses Projekt wurde hauptsächlich an wenigen Wochenenden entwickelt und ist noch in einem grundlegenden Stadium. Meine zeitlichen Ressourcen unter der Woche sind begrenzt, was die schnelle Implementierung komplexerer Features erschwert hat.

---

### Technologischer Stack

Das Projekt nutzt die folgenden Technologien:

- **Odin DNS (Backend):** Go
- **Odin DNS Manager (Frontend):** SvelteKit
- **Persistente Datenhaltung:** MySQL
- **Caching:** Redis
- **Metriken:** ClickHouse
- **Reverse-Proxy & TLS:** Caddy (mit automatischer Let's Encrypt Integration)
