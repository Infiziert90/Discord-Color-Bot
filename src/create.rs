use image;

pub fn create_image(color: &u64) -> Vec<u8> {
    let red = ((color >> 16) & 0xff) as u8;
    let green = ((color >> 8) & 0xff) as u8;
    let blue = (color & 0xff) as u8;

    let mut imgbuf = image::ImageBuffer::new(200, 100);
    for (_, _, pixel) in imgbuf.enumerate_pixels_mut() {
        *pixel = image::Rgb([red, green, blue]);
    }

    let mut bytes: Vec<u8> = Vec::new();
    let mut cursor = std::io::Cursor::new(&mut bytes);
    imgbuf.write_to(&mut cursor, image::ImageOutputFormat::Png).unwrap();

    bytes
}
