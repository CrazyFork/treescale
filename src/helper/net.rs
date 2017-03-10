#![allow(dead_code)]

use std::mem;

/// helper functions for network operations
pub struct NetHelper {
}

impl NetHelper {
    /// Converting u32 integer to BigEndian bytes
    /// Returns 0 if it is unable to make it
    /// Returns final offset in buffer after adding bytes to it
    pub fn u32_to_bytes(number: u32, buffer: &mut Vec<u8>, offset: usize) -> usize {
        if buffer.len() - offset < 4 {
            return 0;
        }

        let endian_bytes = unsafe {
            mem::transmute::<u32, [u8; 4]>(number.to_be())
        };

        buffer[offset + 0] = endian_bytes[0];
        buffer[offset + 1] = endian_bytes[1];
        buffer[offset + 2] = endian_bytes[2];
        buffer[offset + 3] = endian_bytes[3];

        offset + 4
    }

    /// Converting u64 integer to BigEndian bytes
    /// Returns 0 if it is unable to make it
    /// Returns final offset in buffer after adding bytes to it
    pub fn u64_to_bytes(number: u64, buffer: &mut Vec<u8>, offset: usize) -> usize {
        if buffer.len() - offset < 8 {
            return 0;
        }

        let endian_bytes = unsafe {
            mem::transmute::<u64, [u8; 8]>(number.to_be())
        };

        buffer[offset + 0] = endian_bytes[0];
        buffer[offset + 1] = endian_bytes[1];
        buffer[offset + 2] = endian_bytes[2];
        buffer[offset + 3] = endian_bytes[3];
        buffer[offset + 4] = endian_bytes[4];
        buffer[offset + 5] = endian_bytes[5];
        buffer[offset + 6] = endian_bytes[6];
        buffer[offset + 7] = endian_bytes[7];

        offset + 4
    }
}