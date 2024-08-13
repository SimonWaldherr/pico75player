import hub75
import micropython
import struct

# Constants for display size
HEIGHT = 64
WIDTH = 64

# Initialize the Hub75 LED matrix display
display = hub75.Hub75(WIDTH, HEIGHT)

class SAGHeader:
    """Class to represent and parse the SAG file header."""
    def __init__(self, data):
        self.signature = data[0:3].decode('utf-8')
        self.version = data[3]
        self.width, self.height, self.frame_count, self.frame_delay = struct.unpack('>HHHH', data[4:12])
        self.color_palette = [data[i:i+3] for i in range(12, 12 + 768, 3)]  # Extract color palette

@micropython.native
def read_sag_header(file):
    """Read and return the SAG header from the file."""
    header_data = file.read(12 + 768)  # Header size is 12 bytes + 768 bytes for the color palette
    return SAGHeader(header_data)

@micropython.native
def read_and_display_line(file, header, prev_line, y):
    """Read and process a single line of a frame from the SAG file."""
    current_line = [None] * header.width

    for x in range(0, header.width, 8):
        identical_byte = ord(file.read(1))  # Read the identical-byte
        pixel_block = file.read(8)          # Read the next 8 pixels
        process_pixel_block(x, y, identical_byte, pixel_block, prev_line, current_line, header.color_palette)
    return current_line

@micropython.native
def process_pixel_block(x, y, identical_byte, pixel_block, prev_line, current_line, palette):
    """
    Process an 8-pixel block and update the display.
    
    This function checks each pixel in the block, uses the identical-byte to determine if the pixel
    should be reused from the previous line or updated with new data. The updated pixel is displayed
    on the Hub75 matrix.
    """
    for bit in range(8):
        #if x + bit >= len(current_line):
        #    break

        if not identical_byte & (1 << (7 - bit)):
        #else:
            # Use the new pixel data
            color_index = pixel_block[bit]
            r, g, b = palette[color_index]
            current_line[x + bit] = (r, g, b)  # Store in the current line
            display.set_pixel(x + bit, y, r, g, b)

@micropython.native
def draw_sag_animation(filename):
    """Draw the SAG animation on the Hub75 matrix, processing line by line and looping endlessly."""
    prev_lines = [[None] * WIDTH for _ in range(HEIGHT)]  # Initialize previous frame's lines

    while True:  # Infinite loop to play the animation endlessly
        with open(filename, 'rb') as file:
            header = read_sag_header(file)  # Read the SAG header

            for frame_index in range(header.frame_count):
                for y in range(header.height):
                    current_line = read_and_display_line(file, header, prev_lines[y], y)  # Process line by line
                    prev_lines[y] = current_line  # Update the previous line for the next iteration
                display.update(display)

def main():
    input_filename = 'output.sag'  # Replace with your SAG file
    display.start()  # Start the Hub75 display
    draw_sag_animation(input_filename)  # Draw the animation on the display

if __name__ == '__main__':
    main()


