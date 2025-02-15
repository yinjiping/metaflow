use bitflags::bitflags;
use hpack::decoder::Decoder;

const STATIC_INDEX_MIN: usize = 1;
const STATIC_INDEX_MAX: usize = 61;

#[derive(PartialEq, Copy, Clone, Debug)]
pub enum ParseError {
    HeaderIndexOutOfBounds,
    NotStaticField,
    InvalidIntger,
    InvalidMaxDynamicSize(u32, u32),
    InvalidInput,
    NotEnoughOctets,
    InvalidHuffmanCode,
}

bitflags! {
    #[derive(Default)]
    pub struct FieldRepresentation:u8 {
        const INDEXED = 128;
        const LITERAL_WITH_INCREMENTAL_INDEXING = 64;
        const SIZE_UPDATE = 32;
        const LITERAL_NEVER_INDEXED = 16;
        const MASK = 0b11110000;
    }
}

pub struct Parser<'a> {
    decoder: Decoder<'a>,
}

fn parse_int(buf: &[u8], prefix: u8) -> Result<(usize, usize), ParseError> {
    if prefix < 1 || prefix > 8 {
        return Err(ParseError::InvalidIntger);
    }

    if buf.len() < 1 {
        return Err(ParseError::NotEnoughOctets);
    }

    let mask = if prefix == 8 {
        0xffu8
    } else {
        (1u8 << prefix).wrapping_sub(1)
    };

    let mut value = (buf[0] & mask) as usize;

    if value < (mask as usize) {
        return Ok((value, 1));
    }

    let mut len = 1;
    let mut shift = 0;

    const OCTET_LIMIT: usize = 5;

    for &b in buf[1..].iter() {
        len += 1;
        value += ((b & 0x7f) as usize) << shift;

        if b & 0x80 != 0x80 {
            return Ok((value, len));
        }

        if len == OCTET_LIMIT {
            return Err(ParseError::InvalidIntger);
        }

        shift += 7;
    }

    Err(ParseError::NotEnoughOctets)
}

impl Parser<'_> {
    pub fn new() -> Parser<'static> {
        Parser {
            decoder: Decoder::new(),
        }
    }

    fn parse_kv_pair(
        &mut self,
        buf: &[u8],
        prefix: u8,
    ) -> Result<(Option<Vec<(Vec<u8>, Vec<u8>)>>, usize), ParseError> {
        let (index, index_len) = parse_int(buf, prefix)?;
        let mut val_len = index_len;

        if index_len > buf.len() {
            return Err(ParseError::InvalidInput);
        }

        if index != 0 {
            // Indexed
            let (str_len, len) = parse_int(&buf[index_len..], 7)?;
            val_len = val_len + str_len + len;
            // RFC7541附录A(https://datatracker.ietf.org/doc/html/rfc7541#appendix-A)规定：
            // 静态表index从1到61，共60项。如果index大于61, 意味着这是一个dynamic table的
            // index，我们无法解出index对应的value，应该跳过对应的字节继续解析。
            if index > STATIC_INDEX_MAX {
                return Ok((None, val_len));
            }
        } else {
            // New Name
            let (name_len, len) = parse_int(&buf[1..], 7)?;
            let key_len = name_len + len;

            if key_len + 1 >= buf.len() {
                return Err(ParseError::InvalidInput);
            }

            let (value_len, len) = parse_int(&buf[(1 + key_len)..], 7)?;
            val_len = value_len + key_len + len + val_len;
        }

        if val_len > buf.len() {
            return Err(ParseError::InvalidInput);
        }

        match self.decoder.decode(&buf[..val_len]) {
            Ok(rst) => Ok((Some(rst), val_len)),
            Err(_) => Err(ParseError::InvalidHuffmanCode),
        }
    }

    fn parse_indexed(
        &mut self,
        buf: &[u8],
    ) -> Result<(Option<Vec<(Vec<u8>, Vec<u8>)>>, usize), ParseError> {
        let (index, index_len) = parse_int(buf, 7)?;

        if index_len > buf.len() {
            return Err(ParseError::InvalidInput);
        }

        if index >= STATIC_INDEX_MIN && index <= STATIC_INDEX_MAX {
            match self.decoder.decode(&buf[..index_len]) {
                Ok(rst) => Ok((Some(rst), index_len)),
                Err(_) => Err(ParseError::InvalidHuffmanCode),
            }
        } else {
            Ok((None, index_len))
        }
    }

    fn parse_indexing(
        &mut self,
        buf: &[u8],
    ) -> Result<(Option<Vec<(Vec<u8>, Vec<u8>)>>, usize), ParseError> {
        self.parse_kv_pair(buf, 6)
    }

    fn parse_sizeup(&mut self, buf: &[u8]) -> Result<usize, ParseError> {
        let (_, consumed) = parse_int(buf, 5)?;
        Ok(consumed)
    }

    fn parse_never_indexed(
        &mut self,
        buf: &[u8],
    ) -> Result<(Option<Vec<(Vec<u8>, Vec<u8>)>>, usize), ParseError> {
        self.parse_kv_pair(buf, 4)
    }

    fn parse_without_indexing(
        &mut self,
        buf: &[u8],
    ) -> Result<(Option<Vec<(Vec<u8>, Vec<u8>)>>, usize), ParseError> {
        self.parse_kv_pair(buf, 4)
    }

    pub fn parse_one_field(
        &mut self,
        input: &[u8],
        offset: usize,
        output: &mut Vec<(Vec<u8>, Vec<u8>)>,
    ) -> Result<usize, ParseError> {
        let field_flag =
            FieldRepresentation::from_bits_truncate(input[offset]) & FieldRepresentation::MASK;
        let consumed = if field_flag.contains(FieldRepresentation::INDEXED) {
            let (parse_rst, len) = self.parse_indexed(&input[offset..])?;
            if let Some(mut value) = parse_rst {
                output.append(&mut value);
            }
            len
        } else if field_flag.contains(FieldRepresentation::LITERAL_WITH_INCREMENTAL_INDEXING) {
            let (parse_rst, len) = self.parse_indexing(&input[offset..])?;
            if let Some(mut value) = parse_rst {
                output.append(&mut value);
            }
            len
        } else if field_flag.contains(FieldRepresentation::SIZE_UPDATE) {
            let len = self.parse_sizeup(&input[offset..])?;
            len
        } else if field_flag.contains(FieldRepresentation::LITERAL_NEVER_INDEXED) {
            let (parse_rst, len) = self.parse_never_indexed(&input[offset..])?;
            if let Some(mut value) = parse_rst {
                output.append(&mut value);
            }
            len
        } else {
            let (parse_rst, len) = self.parse_without_indexing(&input[offset..])?;
            if let Some(mut value) = parse_rst {
                output.append(&mut value);
            }
            len
        };

        Ok(consumed)
    }

    pub fn parse(&mut self, input: &[u8]) -> Result<Vec<(Vec<u8>, Vec<u8>)>, ParseError> {
        let mut header_list = Vec::new();
        let mut offset = 0;

        while offset < input.len() {
            offset += self.parse_one_field(input, offset, &mut header_list)?;
        }

        Ok(header_list)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn parse_field_indexed() {
        let mut parser = Parser::new();

        let buffer1 = [0x82]; // static table index
        let r1 = parser.parse(&buffer1).unwrap();

        assert_eq!(b":method", r1[0].0.as_slice());
        assert_eq!(b"GET", r1[0].1.as_slice());

        let buffer2 = [0xbf]; // dynamic table index
        let r2 = parser.parse(&buffer2).unwrap();

        assert!(r2.is_empty());
        assert_eq!(0, r2.len());
    }

    #[test]
    fn parse_field_incremental_indexing() {
        let mut parser = Parser::new();
        let buffer_indexed = [
            0x50, 0x8d, 0x9b, 0xd9, 0xab, 0xfa, 0x52, 0x42, 0xcb, 0x40, 0xd2, 0x5f, 0xa5, 0x23,
            0xb3,
        ];
        let buffer_new_index = [
            0x40, 0x92, 0xb6, 0xb9, 0xac, 0x1c, 0x85, 0x58, 0xd5, 0x20, 0xa4, 0xb6, 0xc2, 0xad,
            0x61, 0x7b, 0x5a, 0x54, 0x25, 0x1f, 0x81, 0x0f,
        ];

        let r1 = parser.parse(&buffer_indexed).unwrap();

        assert_eq!(b"accept-encoding", r1[0].0.as_slice());
        assert_eq!(b"gzip, deflate, br", r1[0].1.as_slice());

        let r2 = parser.parse(&buffer_new_index).unwrap();
        assert_eq!(b"upgrade-insecure-requests", r2[0].0.as_slice());
        assert_eq!(b"1", r2[0].1.as_slice());
    }

    #[test]
    fn parse_field_without_indexing() {
        let mut parser = Parser::new();

        let buffer_indexed = [
            0x04, 0x9a, 0x62, 0x43, 0x91, 0x8a, 0x47, 0x55, 0xa3, 0xa1, 0x89, 0xd3, 0x4d, 0x0c,
            0x44, 0x84, 0x8d, 0x26, 0x23, 0x04, 0x42, 0x18, 0x4c, 0xe5, 0xa4, 0xab, 0x91, 0x08,
        ];

        let buffer_indexed2 = [0x0f, 0x0d, 0x02, 0x34, 0x36];

        let buffer_new_index = [
            0x00, 0x90, 0x21, 0xea, 0x49, 0x6a, 0x4a, 0xc8, 0x29, 0x2d, 0xb0, 0xc9, 0xf4, 0xb5,
            0x67, 0xa0, 0xc4, 0xf5, 0xff, 0xbd, 0x05, 0x41, 0x2c, 0x35, 0x69, 0x59, 0x16, 0x11,
            0x4f, 0xf4, 0x82, 0xd1, 0x2f, 0xfa, 0x53, 0xe5, 0x7a, 0x4f, 0xec, 0xd4, 0x50, 0x35,
            0xea, 0x2a, 0x54, 0xf9, 0x5e, 0x93, 0xfb, 0x35, 0x14, 0x0d, 0x73, 0xd9, 0x32, 0x9f,
            0x2b, 0xd2, 0x7f, 0x66, 0xa2, 0x81, 0xae, 0x43, 0xd2, 0xa7, 0xfa, 0xb6, 0xa4, 0x0e,
            0x52, 0xac, 0x6a, 0xa8, 0x35, 0x45, 0xff, 0x4a, 0x7f, 0xab, 0x6a, 0x40, 0xe5, 0x2a,
            0xc5, 0xee, 0x3a, 0x3f, 0xd2, 0x9f, 0x2b, 0x9e, 0xb4, 0x9a, 0x93, 0x7b, 0x2d, 0x1e,
            0x97, 0x21, 0xe9, 0x50, 0xf5, 0xa4, 0xd4, 0x9b, 0xd9, 0x68, 0xf4, 0xba, 0x19, 0x5c,
            0x74, 0x8f, 0xd9, 0xea, 0x1f, 0x84, 0x2e, 0x43, 0xd2, 0xa7, 0x8f, 0x1e, 0x17, 0x98,
            0xe7, 0x9a, 0x82, 0xa4, 0x73, 0x52, 0x3a, 0x87, 0x31, 0x6c, 0x5c, 0x87, 0xa5, 0x4f,
            0x1e, 0x3c, 0x2f, 0x31, 0xcf, 0x35, 0x05, 0x58, 0x75, 0x0e, 0x8f, 0x49, 0x31, 0x10,
            0xb9, 0x0f, 0x4a, 0x89, 0x1c, 0xd4, 0x8e, 0xa1, 0xcc, 0x5b, 0x17, 0x98, 0xe7, 0x9a,
            0x82, 0xae, 0x43, 0xd2, 0xa7, 0x8f, 0x1e, 0x17, 0xf4, 0x7b, 0x53, 0x6c, 0x65, 0x5c,
            0x87, 0xa5, 0x44, 0x2f, 0xe9, 0x26, 0xa6, 0x65, 0xc8, 0x7a, 0x7e, 0xd4, 0x35, 0x33,
            0x2c, 0x8b, 0x08, 0xa7, 0xfa, 0x41, 0x68, 0x97, 0xfd, 0x29, 0xf2, 0xbd, 0x27, 0xf6,
            0x6a, 0x28, 0x1a, 0xf5, 0x15, 0x2a, 0x7c, 0xaf, 0x49, 0xfd, 0x9a, 0x8a, 0x06, 0xb9,
            0xec, 0x99, 0x4f, 0x95, 0xe9, 0x3f, 0xb3, 0x51, 0x40, 0xd7, 0x21, 0xe9, 0x52, 0x41,
            0xa4, 0x77, 0x14, 0xf9, 0x5c, 0xf5, 0xa4, 0xd4, 0x9b, 0xd9, 0x68, 0xf4, 0xb9, 0x0f,
            0x4a, 0x9e, 0x3c, 0x78, 0x5e, 0x63, 0x9e, 0x6a, 0x0a, 0x91, 0xcd, 0x48, 0xea, 0x1c,
            0xc5, 0xb1, 0x72, 0x1e, 0x95, 0x3c, 0x78, 0xf0, 0xbc, 0xc7, 0x3c, 0xd4, 0x15, 0x61,
            0xd4, 0x3a, 0x3d, 0x24, 0xc4, 0x42, 0xe4, 0x3d, 0x2a, 0x7c, 0xae, 0x93, 0x50, 0x54,
            0x2f, 0x48, 0xeb, 0x8c, 0xfe, 0x57, 0x21, 0xe9, 0x50, 0x75, 0x99, 0x7a, 0x47, 0x5c,
            0x67, 0xf2, 0xb9, 0x0f, 0x4a, 0x84, 0xb0, 0xa3, 0x49, 0xbb, 0x94, 0x87, 0xa6, 0x93,
            0xd4, 0x85, 0xcf, 0x64, 0xca, 0x0e, 0x45, 0xe4, 0x3d, 0xb1, 0xd0, 0x52, 0x50, 0x62,
            0x75, 0x5e, 0xa2, 0xa7, 0xed, 0x49, 0x0b, 0x28, 0xed, 0xa1, 0x2b, 0x22, 0xc2, 0x29,
            0xfe, 0x90, 0x5a, 0x25, 0xff, 0x4a, 0x7c, 0xaf, 0x49, 0xfd, 0x9a, 0x8a, 0x06, 0xbd,
            0x45, 0x4a, 0x9f, 0x2b, 0xd2, 0x7f, 0x66, 0xa2, 0x81, 0xae, 0x7b, 0x26, 0x53, 0xe5,
            0x7a, 0x4f, 0xec, 0xd4, 0x50, 0x35, 0xc8, 0x7a, 0x7e, 0xd4, 0x96, 0xc1, 0xd2, 0x55,
            0x91, 0x61, 0x14, 0xf9, 0x5c, 0xf5, 0xa4, 0xd4, 0x9b, 0xd9, 0x68, 0xf4, 0xb9, 0x0f,
            0x4a, 0x9e, 0x3c, 0x78, 0x5e, 0x63, 0x9e, 0x6a, 0x0a, 0x91, 0xcd, 0x48, 0xea, 0x1c,
            0xc5, 0xb1, 0x72, 0x1e, 0x95, 0x3c, 0x78, 0xf0, 0xbc, 0xc7, 0x3c, 0xd4, 0x15, 0x61,
            0xd4, 0x3a, 0x3d, 0x24, 0xc4, 0x42, 0xe4, 0x3d, 0x2a, 0x78, 0xf1, 0xe1, 0x7f, 0x47,
            0xb5, 0x36, 0xc6, 0x55, 0xaa, 0x39, 0x0e, 0x7e, 0xa6, 0x2a, 0xe4, 0x3d, 0x2a, 0x26,
            0xc1, 0x93, 0xa9, 0x6c, 0x49, 0x50, 0x95, 0xcf, 0x64, 0xca, 0x78, 0xf1, 0xe1, 0x74,
            0x5b, 0x67, 0x72, 0xfa, 0x98, 0xde, 0xe9, 0x3a, 0xe4, 0x3d, 0x2a, 0x0c, 0x84, 0x3d,
            0xb5, 0x25, 0x0b, 0xca, 0x6b, 0x0b, 0x29, 0xfc, 0xae, 0x43, 0xd2, 0xa0, 0xc8, 0x43,
            0xdb, 0x52, 0x50, 0xbc, 0xa6, 0xb0, 0xb2, 0x9f, 0xca, 0xe4, 0x3d, 0x2b, 0x92, 0xa5,
            0x3c, 0x78, 0xf0, 0xbf, 0xa3, 0xda, 0x9b, 0x63, 0x2a, 0xe4, 0x3d, 0x3f, 0x6a, 0x21,
            0x3e, 0xa8, 0x2a, 0xc8, 0xb0, 0x8a, 0x7f, 0xa4, 0x16, 0x89, 0x7f, 0xd2, 0x9f, 0x2b,
            0xd2, 0x7f, 0x66, 0xa2, 0x81, 0xaf, 0x51, 0x52, 0xa7, 0xca, 0xf4, 0x9f, 0xd9, 0xa8,
            0xa0, 0x6b, 0x9e, 0xc9, 0x94, 0xf9, 0x5e, 0x93, 0xfb, 0x35, 0x14, 0x0d, 0x72, 0x1e,
            0x95, 0x3f, 0xd5, 0xb5, 0x20, 0x72, 0x95, 0x63, 0x55, 0x41, 0xaa, 0x2f, 0xfa, 0xfb,
            0x50, 0x87, 0xaa, 0xa2, 0x91, 0x2b, 0x22, 0xc2, 0x29, 0xfe, 0x90, 0x5a, 0x25, 0xff,
            0x4a, 0x7c, 0xaf, 0x49, 0xfd, 0x9a, 0x8a, 0x06, 0xbd, 0x45, 0x4a, 0x9f, 0x2b, 0xd2,
            0x7f, 0x66, 0xa2, 0x81, 0xae, 0x7b, 0x26, 0x53, 0xe5, 0x7a, 0x4f, 0xec, 0xd4, 0x50,
            0x35, 0xc8, 0x7a, 0x54, 0xf9, 0x5c, 0xf5, 0xa4, 0xd4, 0x9b, 0xd9, 0x68, 0xf4, 0xb9,
            0x0f, 0x4a, 0x9e, 0x3c, 0x78, 0x5e, 0x63, 0x9e, 0x6a, 0x0a, 0x91, 0xcd, 0x48, 0xea,
            0x1c, 0xc5, 0xb1, 0x72, 0x1e, 0x95, 0x3c, 0x78, 0xf0, 0xbc, 0xc7, 0x3c, 0xd4, 0x15,
            0x61, 0xd4, 0x3a, 0x3d, 0x24, 0xc4, 0x42, 0xe4, 0x3d, 0x2a, 0x7c, 0xae, 0x93, 0x50,
            0x54, 0x2f, 0x48, 0xeb, 0x8c, 0xfe, 0x57, 0x21, 0xe9, 0x50, 0x75, 0x99, 0x7a, 0x47,
            0x5c, 0x67, 0xf2, 0xb9, 0x0f, 0x4f, 0xda, 0x84, 0x9c, 0xd4, 0x48, 0xb2, 0x2c, 0x22,
            0x9f, 0x2b, 0x9e, 0xb4, 0x9a, 0x93, 0x7b, 0x2d, 0x1e, 0x97, 0x21, 0xe9, 0x53, 0xc7,
            0x8f, 0x0b, 0xcc, 0x73, 0xcd, 0x41, 0x52, 0x39, 0xa9, 0x1d, 0x43, 0x98, 0xb6, 0x2e,
            0x43, 0xd2, 0xa7, 0x8f, 0x1e, 0x17, 0x98, 0xe7, 0x9a, 0x82, 0xac, 0x3a, 0x87, 0x47,
            0xa4, 0x98, 0x88, 0x5c, 0x87, 0xa5, 0x4f, 0x1e, 0x3c, 0x2f, 0xe8, 0xf6, 0xa6, 0xd8,
            0xca, 0xb5, 0x47, 0x21, 0xcf, 0xd4, 0xc5, 0x5c, 0x87, 0xa5, 0x44, 0xd8, 0x32, 0x75,
            0x2d, 0x89, 0x2a, 0x12, 0xb9, 0xec, 0x99, 0x4f, 0x1e, 0x3c, 0x2e, 0x8b, 0x6c, 0xee,
            0x5f, 0x53, 0x1b, 0xdd, 0x27, 0x5c, 0x87, 0xa5, 0x41, 0x90, 0x87, 0xb6, 0xa4, 0xa1,
            0x79, 0x4d, 0x61, 0x65, 0x3f, 0x95, 0xc8, 0x7a, 0x54, 0x19, 0x08, 0x7b, 0x6a, 0x4a,
            0x17, 0x94, 0xd6, 0x16, 0x53, 0xf9, 0x5c, 0x87, 0xa5, 0x72, 0x54, 0xa7, 0x8f, 0x1e,
            0x17, 0xf4, 0x7b, 0x53, 0x6c, 0x65, 0x5c, 0x87, 0xa7,
        ];

        let r1 = parser.parse(&buffer_indexed).unwrap();

        assert_eq!(b":path", r1[0].0.as_slice());
        assert_eq!(
            b"/doc/manual/html/_static/css/theme.css",
            r1[0].1.as_slice()
        );

        let r2 = parser.parse(&buffer_new_index).unwrap();

        assert_eq!(b"content-security-policy", r2[0].0.as_slice());
        assert_eq!(1, r2.len());

        let r3 = parser.parse(&buffer_indexed2).unwrap();

        assert_eq!(b"content-length", r3[0].0.as_slice());
        assert_eq!(b"46", r3[0].1.as_slice());
    }

    #[test]
    fn parse_http2_headers_block() {
        let mut parser = Parser::new();
        let buffer1 = [
            0x82, 0x05, 0x97, 0x60, 0xb5, 0x2d, 0xc3, 0x73, 0x12, 0x9a, 0xc2, 0xca, 0x7f, 0x2c,
            0x36, 0x25, 0xc0, 0xb8, 0x58, 0x94, 0xd6, 0x21, 0x36, 0x5b, 0x53, 0x1f, 0x41, 0x8c,
            0xf1, 0xe3, 0xc2, 0xf4, 0x9f, 0xd9, 0xa8, 0xa0, 0x6b, 0x9e, 0xc9, 0xbf, 0x87, 0x7a,
            0xb3, 0xd0, 0x7f, 0x66, 0xa2, 0x81, 0xb0, 0xda, 0xe0, 0x53, 0xfa, 0xfc, 0x08, 0x7e,
            0xd4, 0xce, 0x6a, 0xad, 0xf2, 0xa7, 0x97, 0x9c, 0x89, 0xc6, 0xbe, 0xd4, 0xb3, 0xbd,
            0xc6, 0xc4, 0xb8, 0x3f, 0xb5, 0x31, 0x14, 0x9d, 0x4e, 0xc0, 0x80, 0x10, 0x00, 0x20,
            0x0a, 0x98, 0x4d, 0x61, 0x65, 0x3f, 0x96, 0x1b, 0x12, 0xe0, 0x53, 0xb0, 0x49, 0x7c,
            0xa5, 0x89, 0xd3, 0x4d, 0x1f, 0x43, 0xae, 0xba, 0x0c, 0x41, 0xa4, 0xc7, 0xa9, 0x8f,
            0x33, 0xa6, 0x9a, 0x3f, 0xdf, 0x9a, 0x68, 0xfa, 0x1d, 0x75, 0xd0, 0x62, 0x0d, 0x26,
            0x3d, 0x4c, 0x79, 0xa6, 0x8f, 0xbe, 0xd0, 0x01, 0x77, 0xfe, 0xbe, 0x58, 0xf9, 0xfb,
            0xed, 0x00, 0x17, 0x7b, 0x51, 0x8b, 0x2d, 0x4b, 0x70, 0xdd, 0xf4, 0x5a, 0xbe, 0xfb,
            0x40, 0x05, 0xdb, 0x50, 0x8d, 0x9b, 0xd9, 0xab, 0xfa, 0x52, 0x42, 0xcb, 0x40, 0xd2,
            0x5f, 0xa5, 0x23, 0xb3, 0x40, 0x92, 0xb6, 0xb9, 0xac, 0x1c, 0x85, 0x58, 0xd5, 0x20,
            0xa4, 0xb6, 0xc2, 0xad, 0x61, 0x7b, 0x5a, 0x54, 0x25, 0x1f, 0x81, 0x0f,
        ];

        let r1 = parser.parse(&buffer1).unwrap();

        assert_eq!(9, r1.len());
        assert_eq!(b":method", r1[0].0.as_slice());
        assert_eq!(b"GET", r1[0].1.as_slice());
        assert_eq!(b"user-agent", r1[4].0.as_slice());
        assert_eq!(
            b"Mozilla/5.0 (X11; Linux x86_64; rv:52.0) Gecko/20100101 Firefox/52.0",
            r1[4].1.as_slice()
        );
        assert_eq!(b"upgrade-insecure-requests", r1[8].0.as_slice());
        assert_eq!(b"1", r1[8].1.as_slice());

        let buffer2 = [
            0x42, 0x03, 0x50, 0x55, 0x54, 0x04, 0xb3, 0x62, 0xa1, 0xda, 0x89, 0x56, 0x1d, 0xa9,
            0x9d, 0x8e, 0xe1, 0x62, 0xd2, 0xac, 0x3b, 0x53, 0x39, 0x6a, 0x49, 0x88, 0x34, 0x98,
            0xf5, 0x21, 0x83, 0x52, 0x83, 0x2c, 0xd3, 0x80, 0x00, 0x80, 0xc8, 0x02, 0x00, 0x04,
            0x00, 0x0b, 0x0d, 0xcc, 0xb0, 0xfa, 0x8d, 0x62, 0x1e, 0xa9, 0x4d, 0x65, 0x23, 0x49,
            0x8f, 0x57, 0x86, 0xc1, 0xc0, 0x0f, 0x0d, 0x02, 0x34, 0x36, 0xbf,
        ];

        let r2 = parser.parse(&buffer2).unwrap();

        assert_eq!(4, r2.len());
        assert_eq!(b":method", r2[0].0.as_slice());
        assert_eq!(b"PUT", r2[0].1.as_slice());
        assert_eq!(b"content-length", r2[3].0.as_slice());
        assert_eq!(b"46", r2[3].1.as_slice());
    }
}
