use actix_multipart::Multipart;
use actix_web::{web, App, Error, HttpResponse, HttpServer, Responder};
use futures_util::stream::StreamExt as _;
use serde::{Deserialize, Serialize};
use std::fs::{self, File};
use std::io::Write;
use std::path::Path;

#[derive(Serialize, Deserialize)]
struct Bucket {
    name: String,
}

async fn hello_world() -> impl Responder {
    HttpResponse::Ok().body("hello world")
}

async fn upload_file(
    mut payload: Multipart,
    query: web::Query<Bucket>,
) -> Result<HttpResponse, Error> {
    while let Some(item) = payload.next().await {
        let mut field = item?;
        let content_disposition = field.content_disposition().unwrap();
        let filename = content_disposition.get_filename().unwrap();
        let filepath = format!("./{}/{}", query.name, sanitize_filename::sanitize(filename));

        let f = web::block(|| File::create(filepath)).await??;

        let f = std::sync::Arc::new(std::sync::Mutex::new(f));

        while let Some(chunk) = field.next().await {
            let data = chunk.unwrap();

            let f = std::sync::Arc::clone(&f);

            web::block(move || {
                let mut file = f.lock().unwrap();
                file.write_all(&data).map(|_| ())
            })
            .await??;
        }
    }

    Ok(HttpResponse::Ok().body("File uploaded successfully"))
}

async fn create_bucket(bucket: web::Json<Bucket>) -> impl Responder {
    let dir_name = &bucket.name;
    fs::create_dir_all(dir_name).unwrap();
    HttpResponse::Ok().json(&bucket.name)
}

async fn get_files(query: web::Query<Bucket>) -> impl Responder {
    let dir_name = &query.name;
    if !Path::new(dir_name).exists() {
        return HttpResponse::BadRequest().body("Directory does not exist");
    }

    let files: Vec<String> = fs::read_dir(dir_name)
        .unwrap()
        .map(|res| res.unwrap().path().display().to_string())
        .collect();

    HttpResponse::Ok().json(files)
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    HttpServer::new(|| {
        App::new()
            .route("/", web::get().to(hello_world))
            .route("/uploaded", web::post().to(upload_file))
            .route("/create", web::post().to(create_bucket))
            .route("/list", web::get().to(get_files))
    })
    .bind(("127.0.0.1", 8080))?
    .run()
    .await
}
