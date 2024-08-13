import struct

class SAGHeader:
    def __init__(self, data):
        self.signature = data[0:3].decode('utf-8')
        self.version = data[3]
        self.width, self.height, self.frame_count, self.frame_delay = struct.unpack('>HHHH', data[4:12])
        self.color_palette = [data[i:i+3] for i in range(12, 12 + 768, 3)]

def read_sag_file(filename):
    with open(filename, 'rb') as file:
        header_data = file.read(12 + 768)  # Header size is 12 bytes + 768 bytes for the color palette
        header = SAGHeader(header_data)
        frames = read_frames(file, header)
        return header, frames

def read_frames(file, header):
    frames = []
    prev_frame = None
    for frame_index in range(header.frame_count):
        frame = []
        for y in range(header.height):
            row = []
            for x in range(0, header.width, 8):
                identical_byte = ord(file.read(1))
                pixel_block = file.read(8)
                row.extend(process_pixel_block(x, y, identical_byte, pixel_block, prev_frame, header.color_palette, frame_index))
            frame.append(row)
        frames.append(frame)
        prev_frame = frame
    return frames

def process_pixel_block(x, y, identical_byte, pixel_block, prev_frame, palette, frame_index):
    pixels = []
    for bit in range(8):
        if x + bit >= len(pixel_block):
            break

        if identical_byte & (1 << (7 - bit)):
            if prev_frame:
                r, g, b = prev_frame[y][x + bit]
        else:
            color_index = pixel_block[bit]
            r, g, b = palette[color_index]

        pixels.append((r, g, b))
        print(f"Frame: {frame_index} X: {x + bit} Y: {y} R: {r} G: {g} B: {b}")
    return pixels

def main():
    input_filename = 'output.sag'  # Replace with your SAG file
    header, frames = read_sag_file(input_filename)

if __name__ == '__main__':
    main()
