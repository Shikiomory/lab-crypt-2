# Запуск проекта

## Локальный запуск

### Зависимости

- Go версии 1.27.3 или новее.
- Исходное изображение в формате PNG.
- Python 3 и пакет `matplotlib` — для построения изображения гистограммы.

Установите `matplotlib`, если он еще не установлен:

```sh
python -m pip install matplotlib
```

Запускайте команды из корня проекта. Для изображения по умолчанию (`imgs/pnoto.png`) выполните:

```sh
go run ./src
python scripts/plot.py
```

Для другого изображения передайте путь к PNG аргументом:

```sh
go run ./src imgs/noise_texture.png
python scripts/plot.py
```

Go-программа встраивает сгенерированный текст в изображение методами PlainLSB, PlainCipherLSB и RandLSB. Для каждого метода проверяется восстановление сообщения и рассчитываются метрики. Она сохраняет стегограммы в `imgs/`, CSV с метриками и данными гистограммы, а также изображение разницы — в `results/`. Команда Python строит `results/histogram.png` по данным из `results/histogram.csv`.

Сообщение встраивается с нагрузкой 50% и 100% от емкости изображения. Дополнительно для гистограммы и карты разницы создается стегограмма RandLSB с нагрузкой 15%.

Вместо аргумента можно задать исходное изображение через `SRC_IMG_PATH`. В PowerShell:

```powershell
$env:SRC_IMG_PATH = "imgs/noise_texture.png"
go run ./src
python scripts/plot.py
```

Для локального запуска переменные путей по умолчанию такие:

| Переменная окружения | Значение по умолчанию | Назначение |
| --- | --- | --- |
| `SRC_IMG_PATH` | `imgs/gradient.png` | Исходное PNG |
| `HISTOGRAM_CSV_PATH` | `results/histogram.csv` | CSV с данными гистограммы |
| `HISTOGRAM_IMG_PATH` | `results/histogram.png` | Изображение гистограммы |

`HISTOGRAM_CSV_PATH` и `HISTOGRAM_IMG_PATH` должны совпадать для Go-программы и `scripts/plot.py`, поскольку программа сохраняет CSV по этому пути, а скрипт использует эти переменные для чтения CSV и сохранения изображения.

Метрики всегда сохраняются в `results/metrics.csv`, а карта разницы — в `results/difference.png`.

### Результаты

Для `photo.png` в каталоге `imgs/` появятся стегограммы:

- PlainLSB: `photo-plainLsb_50-percent-enc.png`, `photo-plainLsb_100-percent-enc.png`;
- PlainCipherLSB: `photo-plainLsbAes_50-percent-enc.png`, `photo-plainLsbAes_100-percent-enc.png`;
- RandLSB: `photo-encryptedRandomLsb_50-percent-enc.png`, `photo-encryptedRandomLsb_100-percent-enc.png`;
- дополнительная стегограмма RandLSB с нагрузкой 15%: `photo-encryptedRandomLsb-enc.png`.

Расшифрование выполняется внутри программы для проверки корректности встраивания. Отдельные файлы с расшифрованными изображениями не создаются; отдельная команда для запуска только встраивания или только проверки не предусмотрена.

## Запуск в Docker

### Зависимости

- Docker и Docker Compose.
- Исходное изображение в формате PNG в каталоге `imgs/`.

### Команды

Запуск с изображением по умолчанию (`imgs/gradient.png`):

```sh
docker compose up --build
```

Запуск с другим изображением, например `imgs/photo.png`:

- **Linux/macOS:**

    ```sh
    SRC_IMG_PATH=/app/imgs/photo.png docker compose up --build
    ```

- **Windows (PowerShell):**

    ```powershell
    $env:SRC_IMG_PATH="/app/imgs/photo.png"; docker compose up --build
    ```

- **Универсально (файл .env рядом с docker-compose.yml)**

    ```env
    SRC_IMG_PATH=/app/imgs/photo.png
    ```

    И запуск стандартной командой:

    ```sh
    docker compose up --build
    ```

### Параметры

| Переменная окружения | Значение по умолчанию | Назначение |
| --- | --- | --- |
| `SRC_IMG_PATH` | `/app/imgs/gradient.png` | Исходное PNG внутри контейнера |
| `HISTOGRAM_CSV_PATH` | `/app/results/histogram.csv` | Путь сохранения данных гистограммы |
| `HISTOGRAM_IMG_PATH` | `/app/results/histogram.png` | Путь сохранения изображения гистограммы |

Каталог `imgs/` и каталог `results/` доступны контейнеру через Docker Compose, поэтому исходное изображение и результаты остаются на хосте. После работы Go-программы контейнер автоматически запускает `scripts/plot.py` для построения гистограммы.

### Реализованные методы

- PlainLSB — встраивание сообщения в младшие биты цветовых каналов без шифрования.
- PlainCipherLSB — шифрование сообщения AES-CTR перед последовательным встраиванием в младшие биты.
- RandLSB — шифрование сообщения AES-CTR и встраивание в псевдослучайные позиции, зависящие от пароля.
