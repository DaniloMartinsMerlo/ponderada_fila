# Ponderada fila

## Descrição

&emsp; Para atender o solicitado na ponderada eu desenvolvi um backend em GO, uma fila utilizando RabbitMQ, um middleware com dois consumers e um banco de dados utilizando Postgres. Além disso criei um teste unitário para o backend e um teste de carga utilizando o K6. Tudo isso foi containerizado utilizando o Docker, então para executar o projeto basta ter o Docker instalado e rodar o comando `docker-compose up` na raiz do repositório.

&emsp; Para mim não fazia sentido colocar o teste de carga juntamente do docker, então ele está separado, para rodar o teste de carga basta entrar na pasta `load-test` e rodar o comando `k6 run k6_load_test.js`.

## Backend

&emsp; O backend foi desenvolvido utilizando o framework Gin, ele possui uma rota POST para receber as requisições e uma rota GET para listar as requisições processadas. A rota POST recebe um JSON com os campos device_id, timestamp, sensor_type, reading_type e value. Antes de publicar na fila, o backend valida e transforma os dados recebidos através da função processarSensor, cujo comportamento depende do campo reading_type:

* Caso seja "analogica", o campo value é convertido de string para float e armazenado em numeric_value. O campo discrete_value é definido como "0". Caso o valor não seja numérico, a requisição é rejeitada com erro.

* Caso seja "discreto", o campo value é mantido como string e armazenado em discrete_value. O campo numeric_value é definido como 0.

&emsp; Caso o reading_type não seja nenhum dos dois valores acima, a requisição também é rejeitada. Se os dados forem válidos, a mensagem transformada é publicada na fila do RabbitMQ. 

&emsp; A rota GET é um pouco diferente, ela lista as mensagens já processadas, ou seja, as que foram consumidas pelo consumer e armazenadas no banco de dados, porém para isso foi necessário criar duas filas, uma onde a requisição de get é enviada para o consumer e outra onde o consumer envia a resposta para o backend, ou seja, o backend envia uma mensagem para a fila de get, o consumer consome essa mensagem, consulta o banco de dados e publica a resposta na fila de resposta, e por fim o backend consome a mensagem da fila de resposta e retorna para o cliente.

&emsp; Optei por diferenciar os tipos de leitura (analógica e discreta) para garantir que os dados sejam armazenados de forma adequada no banco de dados, facilitando consultas e análises futuras. Além disso, a validação dos dados no backend ajuda a garantir a integridade das informações que estão sendo processadas e armazenadas.

## Middleware

&emsp; O middleware é composto por dois consumers que consomem as mensagens das filas do RabbitMQ. O primeiro consumer é responsável por processar as mensagens e armazená-las no banco de dados. Ele lê as mensagens da fila, extrai os dados e os insere na tabela do Postgres. 

&emsp; O segundo consumer é responsável por consumir as mensagens da fila de get, executar a consulta no banco de dados e retornar as informações em uma fila de respostas para que possam ser consumidas pelo backend.

## Banco de dados

&emsp; O banco de dados utilizado é o Postgres, ele possui uma tabela chamada sensor_data com os campos id, device_id, timestamp, sensor_type, reading_type, discrete_value e numeric_value. O campo id é a chave primária e é auto-incrementado. Os outros campos armazenam as informações recebidas pelo backend e processadas pelo middleware.

&emsp; Para subir o banco de dados está sendo utilizado o Docker que sobe uma imagem do postgres e executa o script initi.sql para criar a tabela necessária para o projeto, sendo assim o banco só existe enquanto o container estiver rodando, ou seja, quando o container for parado os dados serão perdidos, como o objetivo é apenas demonstrar as funcionalidades do sistema, não achei necessário manter os dados.

## Docker

&emsp; O Docker é utilizado para containerizar todos os serviços do projeto. O arquivo docker-compose.yml define 5 serviços:

* back: o backend da aplicação, construído a partir do Dockerfile na pasta ./back. Expõe a porta 8088 e só sobe após o banco e o RabbitMQ estarem saudáveis.
* consumer_post: consumer responsável por processar as mensagens de escrita, construído a partir de ./middleware/post. Também aguarda o banco e o RabbitMQ estarem prontos antes de iniciar.
* consumer_get: consumer responsável por processar as requisições de leitura, construído a partir de ./middleware/get. Segue a mesma condição de inicialização dos demais.
* db: banco de dados PostgreSQL configurado com usuário, senha e nome do banco via variáveis de ambiente. Possui um volume para persistir os dados e um healthcheck que verifica se o banco está aceitando conexões antes de liberar os outros serviços.
* rabbitmq: fila de mensagens utilizada para comunicação entre o backend e os consumers. Expõe as portas 5672 para conexão e 15672 para o painel de gerenciamento. Os logs são persistidos na pasta ./rabbitmq do projeto, e um healthcheck garante que o serviço esteja operacional antes dos demais subirem.

&emsp; O uso do depends_on com condition: service_healthy garante que o backend e os consumers só iniciem após o banco e o RabbitMQ estarem completamente prontos, evitando falhas de conexão na inicialização.

## Testes

&emsp; Como era esperado para a ponderada eu desenvolvi dois testes, um teste unitário para o backend e um teste de carga utilizando o K6. O teste unitário foi desenvolvido utilizando o framework de testes do Go, ele testa a função processarSensor do backend, verificando se os dados são processados corretamente de acordo com o tipo de leitura (analógica ou discreta) e se as validações estão funcionando como esperado. O teste de carga foi desenvolvido utilizando o K6, ele simula múltiplas requisições para o backend para verificar o desempenho do sistema sob grandes quantidades de requisições.

### Uso de I.A.

&emsp; Para o desenvolvimento dessa ponderada eu apenas utilizei inteligências artificiais para me ajudar a resolver problemas que tive na hora de desenvolver os testes unitários. Como eu tive alguns problemas durante a aula onde esse conceito foi passado eu tive algumas dificuldades em simular requisições http para o backend, e a I.A. me ajudou a entender melhor como fazer isso utilizando o framework de testes do Go e o comando `gin.SetMode(gin.TestMode). Fora isso, todo o desenvolvimento do projeto foi feito por mim, sem a ajuda de I.A. para escrever código ou tomar decisões de arquitetura.