# RodriTV

Es un producto destinado a la retransmision de contenido audiovisual expuesto de forma gratuita en la web y exponiendolo en una lista iptv que se puede consumir desde una red local sin autorizacion de por medio.

## Panel de control

El sistema debe exponer un panel de control donde se puedan cargar, modificar y borrar canales. Ademas tambien debe permitir los cambios de configuracion como la cantidad de canales que se dejan prendidos hacia atras y hacia adelante del canal que se esta consumiendo actualmente.
La cantidad de canales desde el primero que se encienden, al iniciar el servidor.

Cada canal debe poder soportar un nombre, el logo, una url donde extraer la programacion (opcional), un logo, un numero de orden (arrancando desde el 1) la lista, opcion para desactivar el canal entero y hasta 5 sources.

Cada source de un canal debe soportar un numero de orden arrancando desde el 1, siendo la 1 la source principal. Tambien tener una url donde se da el stream, el tiempo de buffering (configurable), opcion para desactivar el source y debe establecer si requiere un browser o si el sistema tiene un extractor dedicado para la fuente.

El panel cuando es accedido desde una PC debe ser tener una UI similar a vscode, con un menu de iconos a la izquierda, un panel central donde se renderizan las pantallas seleccionadas en el menu y un panel a la derecha donde se renderizan los formularios de edicion de canales, configs y logs.
El panel debe poder ser accedido desde un telefono smartphone android y la ui debe ser responsiva para que funcione correctamente.

## Playlist

El sistema debe exponer un endpoint con una lista iptv m3u, donde se centralicen todos los canales activos. La misma debe ser accesible sin necesidad de autenticacion. Actualmente se va a usar en un TV **LG 50UP7750PSB (webOS)** que se conectara directamente al servidor usando la aplicacino SS-IPTV.
Ya se hicieron las pruebas y el televisor soporta bien los formatos conocidos de video y audio.

## Comportamiento de canales
El streaming de un canal de priorizar la fluidez por sobre la actualidad. Por lo tanto se configurara un buffer de hasta 10 minutos dependendiendo del canal.
Cuando una source se cae, el sistema debe consumir desde la siguiente y mezclar la señal de video sin que el cliente note que hubo una interrupcion. Es por esto que las fuentes de un canal deben estar encendidas todo el tiempo.
Cada canal debe guardar logs de los y metricas de los procesos que corran y deben poder ser expuestos en el panel.
En caso que un canal tenga una sola source y esta se caiga el sistema debe proporcionar una imagen de video informando que hubo un incoveniente en la transmision, y por detras debe intentar volver a reestablecer la transmision automaticamente. Lo mismo aplica si todas las sources del canal estan caidas.
El sistema debe tener una metrica de uptime del canal, si la transmision de la source esta por encima del 95% no es necesario dejar prendidas las demas sources. Basta con transmitir la señal de problemas tecnicos ante una interrupcion.
Si el uptime de la source cae por debajo del 95% se deben encender las demas sources hasta que el uptime global del canal sub por encima del 95%
El sistema debe soportar como formatos de sources tanto privativos si se tiene el extractor dedicado, asi como youtube live y una transmision m3u8.
Siempre teniendo presente que en caso de interrupcion, se deberá mostrar el contenido de problemas tecnicos para que se mantenga la fluidez.
Cuando una source es m3u8 no hace falta extraer nada, solo retransmitir teniendo en cuenta las demas reglas de negocio.
En el caso de que una source sea youtube se debe permitir cargar la opcion de streaming rotativa de ser posible, esto quiere decir que el sistema debe poder descubrir el enlace de transmision en directo actual. Por lo que se debe poder configurar la url del canal de youtube para esta feature. 

En un principio los canales que vamos a configurar son:
- Telefe: 
    - source 1: https://www.mitelefe.com/telefe-en-vivo
    - source 2: https://www.elocho.tv/el-ocho-en-vivo.html
- El Trece: 
    - source 1: https://livetrx01.vodgc.net/eltrecetv/index.m3u8
    - source 2: http://45.184.109.11/auth/xt42Jq5dny8XNA7KtRsl4mWYnzKc6S--sl5BqsDjzYsYefmHRMn_NgT38bpY24V8p78hOfraRP0gJYCaux58Omy8GlHpth9USfbpQikPhp3P-7ETJYFjJjisvLbgKliz3N_FqjRyTNn1J6mp_nPaxBAfN5humtgs8Myzsqa1mBzEU1aF_TsvZ0cwi7tSWMD2X6KhAFSZ464WW7GsOHtTVH1x9KrI2v73-9P59NvQds6d9Q7-HTn7sO8qRgqODzDJZmfC1XDnhkWVP-W82UW5UWT6ku57qfmyVx7H_8we9ZEvDddiB6TnqLAeftOVm24MrZXTfpbKaaBRYSi_a07ZS1OgeOrFlj7s9C7oaFzrh34Kbu1hvdgQFbQIoFY-LN99E9jOg3M-PP8WXlbEYe9UGb3skMXCHiqINALQPqaQlHjL5Cktr62kAcSQjL-JePaMqmSDerRrmIvlUMqbxyJ8NV9xOK1UmgMPHKX_6RpWgPvDcezZnQd23xmahf3jhl_JySuVZNIP2skINDH0SnCaS4UEBQd2kDvokAzwg1nrk1mKMnxJx4UW3quc9QrB6Ic5cOUZDYtoxgidw1MRRc4ar5FnDSc3g1l_CaAjDJfZfeqL2nIfnNKJ0X2CH60tZG8z
    - source 3: https://www.eltrecetv.com.ar/vivo/
- TN: 
    - source 1: https://www.youtube.com/watch?v=cb12KmMMDJA
    - source 2: https://tn.com.ar/envivo/24hs/
- America TV:
    - source 1: https://www.americatv.com.ar/
    - source 2: https://www.youtube.com/watch?v=zcWXboTnous (canal: https://www.youtube.com/@americaenvivo)
- El Nueve:
    - source 1: https://www.youtube.com/watch?v=yGLVhAklLyM (Canal: https://www.youtube.com/@elnueve)
- NeoTv:
    - source 1: https://videostream.shockmedia.com.ar:19360/neotvdigital/neotvdigital.m3u8
    - source 2: https://neotvdigital.com.ar/en-vivo/
- CM:
    - source 1: https://www.cronica.com.ar/elcanaldelamusica/cm-en-vivo/
- Cosmos:
    - source 1: https://tv.mediacp.eu:19360/cosmos/cosmos.m3u8
    - source 2: https://canaletaplus.com:3922/live/cosmoslive.m3u8
- Disney Channel:
    - https://barinaslive1.boldmss.com/DISNEY/DISNEY.isml/DISNEY.m3u8
- ESPN Premium:
    - source 1: http://45.184.109.11/auth/RI6isVNbPzaW-ji1S154OtyInATOcIsJ6fWWs4iJz4_D12qc-H30qdCSHy3KcJscAk-hgI71WBR4DDw5pRIFwVFHUg8uur1O9kd52ZwiW30wCENXxa-U91CJ2oZuIM59rt-lrsMiV2OCk_XdBhded1RYw-Gdkep58yBr8aQCJa_kOP8IeUWv5-_2Z1fXIWgVsl_CmqhG-8gyk6yx4UhYNRDtpZcS3nHfTyY4l8lnXWVFIFdIHkG7zW-vUyuOB1vua3BlcerKdKmWAw_rF2Mzuj5N1B2g0GB8KIbaT9cykbnyEgJRaGN8SSww9n_nFWw3OW0ApWo4U3UQ95rSndH7MTNsoGfTf5AepBrkJQoW0MpLJGhMLukew0JiSgcliFnaojKwR0x7_TQJlNnIPd9v6oEfBfjtAsQkpzgHSyg-4ZoSic5mdrpMx3x30nzu9aQ2NOECM7G7w3whmlYlmR4inQige05BWeyIR4H_XKwNv_esD9hOl2R45RG48y-9vj4H4tqUDvwh1FZnk18Aq209D1jkWFpI1ZCI_RS2JS1P9hTf-9QtyQInvTjuhrrSJ-ZEbMJhvbiS2BKZGe0xh_8rGGZ7HBwTsdHfGVWhTBmmPWU
- TNT Sports:
    - source 1: http://45.184.109.11/auth/FtuiWwnbNhiXJZ69UNdAr_gbnd54pu_iNFzUUR49oNKt9Ur9xoXAvkk1iuvwkR-kWCQ0xIRWeF5atblIoUo5LIOvIOYLKNFzd6z2AWa5qA8L6DIMcPgtog6MLbuKHUnQWzsrnjmshRvsr3-jyIlfN9oJm3tZrWs1XbuN0DOI09GVxt3pwoS48P2MIVMpDEGChicbihjc2MBOU3rPrTx3bObjn7YbLYN_DvgOfjUH1cGr_XXZkd7ps7sa5QOfiGN-CDubbJzoLNjQoUquvY-52j_xbKQvqKZOYD5oB1aJdrZfa-DexDFOU-rcTLtTEU2tF5Vn2LFX910AFSIph70ujJp390NIWR7IAJjtzTtGBS5hsCOF2h0JI6hKPPmS2GGedZUaC1Fsg7e2CLnv0x7hU0QVhqIG7EpNH3QlosW8oFWXTU7-NBaL1l44FaLZOdLr3Lu5llSfiLcj360jUmV0mvvWjYcxPYxENGvRNwba-IIo8CufcipVxnXNnUSbX-YUUeWuBEFDrtEwGgh6xBEzl3bVM57akN4gKQ73dyvT36T-7KW87wnLFHAPEBDjGaTWxvxItvd385FEdu6WYDE7hYEY5XMXWyXIY560gA-klfk
- Fox Sports:
    - source 1: http://45.184.109.11/auth/P7uSOWh4796q7KXTF86Y611zo_K3InyopWj9pgUVOSEwXnL4WPOH_Cqyj6XXX4AG6ECuWlH3DEpgocbkC4AiPPgHVtus_2z5oj4q6cM_EZKD4dP_SZDKwwXewKxyGSsjNRLAVthRau850XIJEWAm8eybYWKtkf9mriXtMmTk9JY5KypDqgUl-mAzZTiO5BMvoqlxSpNt5A2IZq8W3ENhTN9JjNTqYJvPruvq0S_LYkG9KnGjl8cxyNCnDNETQWZU3e1wZz39r3CtdI03wa55k-7iXs-1CJLLkZp6JQ549PRoYdHc7axycKPeDkngThQly0VEKzgAVr-U-x_vt4odWrXSxFS3hj8gi4XHuZ5Icz6uAeGF65ZVh5oqelIDM5EL2m3pVrGXJFetRCtcf1TfUcvamLQUQR7tpz8S4NaXnKJ_UVmZtpX8a3hf4nAYD5PIztt2rin_bEfAyC4257uCthjaA70B-x6G9eYPDvtKLquyGgj4CGTPxpGK5rT0VVnyI4w2r8LsS-ViBYTBgP5JAnCgD4_IaMRuAaD8nOLnfqvLgHLUpgduA2Sen5Vgqp2_jlBKbr_ImuN1Q4zK_Ggqbq8_PZwoeUipSoXrEvGki02RBDgWbGf9T9R_eOPTdjZn
- Tyc Sports:
    - source 1: https://amg26268-amg26268c14-freelivesports-emea-10267.playouts.now.amagi.tv/ts-us-e2-n2/playlist/amg26268-sportsstudio-tycsports-freelivesportsemea/playlist.m3u8
    - source 2: http://45.170.130.224:8000/play/a020/index.m3u8