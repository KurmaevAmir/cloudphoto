# cloudphoto

## Порядок сборки приложения

На компьютере должны стоять:

- Go версии 1.22 или новее;
- Terraform версии 1.5 или новее;
- yc;
- git.

Перед этим должны быть готовы: 
- DNS-зона `vvotNN.itiscl.ru` в Cloud DNS;
- сертификат Let's Encrypt на `vvotNN.itiscl.ru` и `*.vvotNN.itiscl.ru`.

### Переменные окружения

Для Terraform нужна одна переменная:

- `YC_TOKEN` — токен для входа в Yandex Cloud. Получить его: `export YC_TOKEN=$(yc iam create-token)`.

Для программы cloudphoto нужны три переменные:

- `BUCKET` — имя бакета `photos.vvotNN.itiscl.ru`;
- `AWS_ACCESS_KEY_ID` — идентификатор статического ключа;
- `AWS_SECRET_ACCESS_KEY` — секретная часть статического ключа.

Ключ создаёт Terraform, поэтому все три значения берутся из его вывода (команды ниже).

### Запуск инфраструктуры
### Сборка программы в папке `terraform/` и принимает пять входных переменных:

- `folder_id` — ID каталога в Yandex Cloud;
- `bucket_name` — имя бакета `photos.vvotNN.itiscl.ru`;
- `domain` — домен, куда добавляется DNS-запись `vvotNN.itiscl.ru`;
- `cert_id` — ID сертификата в Certificate Manager;
- `dns_zone_id` — ID DNS-зоны.

Их лучше расположить в `terraform/terraform.tfvars`, образец лежит в
`terraform.tfvars.example`.

Дальше:

```bash
cd terraform
export YC_TOKEN=$(yc iam create-token)
terraform init
terraform apply
```

```bash
B=$(terraform output -raw bucket) && \
K=$(terraform output -raw access_key) && \
S=$(terraform output -raw secret_key) && \
[ -n "$B" ] && [ -n "$K" ] && [ -n "$S" ] && \
printf 'export BUCKET=%s\nexport AWS_ACCESS_KEY_ID=%s\nexport AWS_SECRET_ACCESS_KEY=%s\n' "$B" "$K" "$S" > .cloudphoto.env && \
chmod 600 .cloudphoto.env

source .cloudphoto.env
```

### Сборка программы

Все команды из корня репозитория.

```bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o dist/cloudphoto .
```

ИЛИ

```bash
go build -o cloudphoto .
```

## Использование

```bash
cloudphoto list [--album ALBUM]
cloudphoto upload --album ALBUM [--path PHOTOS_DIR]
cloudphoto download --album ALBUM [--path PHOTOS_DIR]
cloudphoto delete --album ALBUM [--photo PHOTO]
cloudphoto mksite
```

Если `--path` не указан, берётся текущая папка. 
Фотографиями считаются файлы, которые заканчиваются на `.jpg` или `.jpeg`.

## Как устроено хранилище

Схема ключей следующая:

```text
photos/<альбом>/          пустой объект
photos/<альбом>/<файл>    сама фотография, имя файла как было у исходного
index.html                главная страница сайта
error.html                страница ошибки
album1.html … albumN.html страницы альбомов
```

### Обоснование

`photos/`. Так фотографии не смогут перепутаться со страницами сайта. 
Страницы лежат в корне, а фотографии находятся только внутри `photos/`. Если 
назвать альбом `index.html`, его ключ будет `photos/index.html/`, и с настоящей 
главной он не пересечётся.

Альбом является папка в ключе. Так из ключа сразу понятно, к какому альбому
относится фотография и как она называлась. У каждой фотографии ровно один ключ,
значит, она лежит ровно в одном альбоме.

Пустой объект `photos/<альбом>/`. Так как в Object Storage нет настоящих папок,
есть только ключи. Если удалить из альбома все фотографии, то без отдельного 
объекта альбом бы просто исчез. Таким образом, по заданию альбом должен 
остаться пустым. Пустой объект-маркер является решением: `upload` всегда его 
создаёт, а `delete --photo` его не трогает. Удаляется он только вместе со 
всем альбомом через `delete --album`.

В имени нельзя использовать `/`. Слеш разделяет части ключа. Если бы он был в 
имени альбома, ключ можно было бы прочитать двумя способами. Запретить проще и 
понятнее, поэтому на такое имя программа отвечает "Ошибка в аргументах". 
В имени файла `/` и так быть не может.

Имена не перекодируются. Ключи в Object Storage — это обычные строки в UTF-8, 
с учётом регистра. Кириллица, пробелы, `+` и прочее хранятся как есть. 
Поэтому `IMG.JPG` и `img.jpg` разные фотографии, и регистр в именах не 
теряется.

Повторный запуск ничего не меняет. Загрузка перезаписывает объект по тому же 
ключу. Если такая фотография уже была, она просто заменится, дублей не будет. 
`mksite` каждый раз пересобирает все страницы и удаляет лишние `albumN.html`, 
если альбомов стало меньше.

Порядок в списках. Имена альбомов и фотография сортируются по байтам UTF-8,
как Object Storage сортирует ключи. Альбомы при этом сортируются по самим 
именам.