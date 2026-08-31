Quiero que actues como un desarrollador senior en golang y experto en clean architecture y basado en eso hagas lo siguiente:

1. Vas a crear una API REST profesional con las mejores practicas siguiendo la siguiente estructura:
    - Las dependencias van hacia adentro, el domain no debe importar nada de capas externas (usecase, repository, handlers, ni librerias)
    - Desacoplamiento como norma, vamos a depender de abstracciones y no de implementaciones, esto quiero decir que los usecases tendran se comunicaran con los repository por medio de interfaces
    - Inyeccion de dependencias: se deben resolver las dependencias en el archivo main.go

2. Vas a completar los CRUD completos de las dos entidades que estan definidas dentro de la carpeta domain, para actualizar via endpoints utiliza `PATCH`

3. Estandariza las respuestas http con un message dependiendo del caso y metodo que sea llamado, data con la data recuperada en db, error en caso de existir

4. El metodo GetAll debe tener paginacion y la response debe tener un campo meta con el total de registros la pagina en la que estas y las que faltan

5. Crea un archivo docker con el que se pueda levantar la api y conectarla con una db en postgresql

6. Estandariza tambien las response de error con el estatus http adecuado: 404 para lo que no existe, 401 unauthorized, 400 bad request

7. Las response exitosas tambien con el estatus adecuado, 201 para los created, 204 no content para los delete, 200 para los get y el update utiliza lo que consideres una mejor practica 204 no content o 200 regresando el recurso actualizado, te dejo libertad en esto

Manos a la obra.


Quiero que crees un archivo agents.md que le diga a los agentes cual es la estructura que seguimos para trabajar este proyecto y que no tenga que repetirme en cada prompt

Ahora vas a crear la entidad Role y User con los siguientes campos y el crud completo:

1. Role: 
 - id int autoincrement
 - name string (unique) ADMIN, CUSTOMER, EDITOR
2. User:
 - id int autoincrmeent
 - email string (unique)
 - password string encriptada
 - username string (unique)
 - role relacion con la tabla roles

Necesito que hagas un cambio, quiero manejar los precios de los productos en datos enteros y que los ultimos dos digitos sean los dos decimales y esto se transforme al momento de devolver datos al usuario, ejemplo:
    - 10.84 pasarian a ser 1084 en nuestra DB pero el usuario debe seguir viendo el precio como 10.84

Necesito manejar la sesion de los usuarios con jwt con las siguientes reglas:
 - Vas a guardar el username y el role en la info del jwt
 - La duracion debe ser de 30 min
 - Endpoint RefreshToken para el front

Quiero que alteres dos entidades:
 - Product debe tener campo img que sera un string con un link a una imagen del producto, puede ser null
 - User debe tener photo que tambien seria un string al link de una foto de perfil que tambien puede ser null


Quiero que actues como un experto en backend, lenguaje go, clean architecture y design patterns tomando las mejores decisiones para los casos correspondientes:

1. Las entidades product y category solo pueden ser editadas o creadas por usuarios con roles admin o editor, debe guardarse evidencia de que usuario creo o edito cualquier entidad con la fecha

2. La compra de algun producto debe pasar por una api externa de tipo wallet para poder confirmar la compra y agregarsela al usuario, quiero que abstraigas la logica de esta llamada a la api externa, se maneja hasta ahora una peticion tipo POST a la espera de como seria el body y las distintas responses

3. Los usuarios tipo admin y editor pueden comprar productos y recibir siempre un 10% de descuento del precio
