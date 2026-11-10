# US-0020: Pantalla Configuración global
## Descripcion
Implementar la página `GET /config` con el formulario de las 5 claves de SPECS-001.5.3 (`neighbors`, `startup_channels`, `current_channel_ttl_minutes`, `buffer_minutes`, `uptime_threshold`) con atributos de validación de rango en cliente (`min`/`max`/`step`/`required`) y el fragmento `POST /frag/config` que guarda vía `PUT /v1/config` y devuelve confirmación o errores por campo (400) o banner de API no disponible.
## Proyecto
rodritv-web
## Depende de
US-0016, US-0017, US-0018
## Valor
El admin puede ver y modificar la configuración de comportamiento del motor desde el panel, con validación en cliente y errores de rango de la API mostrados junto al campo.
## Checklist:
- [ ] T-0063: Crear página y fragmento de configuración global
