
import java.util.ArrayList;
import java.util.List;

import javax.sound.sampled.*;

class Levels {

    public static void main(String[] args) {
			try {
					new Thread(Levels::captureAndBroadcastAudio).start();
					System.in.read();
			} catch (Exception e) {
					e.printStackTrace();
			} 
    }

    private static TargetDataLine getTargetLineByNameAndFormat(String mixerName, AudioFormat format) {
        for (Mixer.Info mixerInfo : AudioSystem.getMixerInfo()) {
						System.out.println("Checking MixerInfo: "+mixerInfo);
            if (mixerInfo.getName().contains(mixerName) && !mixerInfo.getName().contains("Port")) {	// Exclude ports
                Mixer mixer = AudioSystem.getMixer(mixerInfo);
								System.out.println("Mixer: "+mixer);
                try {
                    DataLine.Info dataLineInfo = new DataLine.Info(TargetDataLine.class, format);
                    if (mixer.isLineSupported(dataLineInfo)) {
                        return (TargetDataLine) mixer.getLine(dataLineInfo);
                    } else {
												System.out.println("Line not supported: " + dataLineInfo);
										}
                } catch (LineUnavailableException e) {
                    e.printStackTrace();
                }
            }
        }
        return null;
    }

		public static float calculateRMS(float[] array) {
			float sumOfSquares = 0.0f;
			for (float num : array) {
					sumOfSquares += num * num;
			}
			float meanOfSquares = sumOfSquares / array.length;
			return (float) Math.sqrt(meanOfSquares);
		}

    private static void captureAndBroadcastAudio() {
        try {
            String name = "E.A.R.S Gain: 18dB";
            AudioFormat format = new AudioFormat(48000, 24, 2, true, true);
						System.out.println("Name: '" + name +"' AudioFormat: "+format);

						try {
								int formatType = FloatSampleTools.getFormatType(format);
							if (formatType == -1) {
								System.out.println("Unsupported format type: " + format);
								return;
							} else {
								System.out.println("Format type: " + FloatSampleTools.formatType2Str(formatType));
							}

						} catch (IllegalArgumentException e) {
							System.out.println("Unsupported sample size: " + format.getSampleSizeInBits());
							return;
						}

            TargetDataLine targetDataLine = getTargetLineByNameAndFormat(name, format);
            if (targetDataLine == null) {
                System.out.println("Name '" + name + "' not found.");
                return;
            }

            targetDataLine.open(format);
            targetDataLine.start();

            byte[] buffer = new byte[targetDataLine.getBufferSize()]; 
						long lastPrintTime = System.currentTimeMillis();

            while (true) {
                int bytesRead = targetDataLine.read(buffer, 0, buffer.length);
                if (bytesRead > 0) {
									float referenceSPL = (float) 94.0;
									int initialCapacity = 2;
									List<float[]> arrayList = new ArrayList<float[]>(initialCapacity);
									for (int i = 0; i < initialCapacity; i++) {
										arrayList.add(new float[bytesRead / format.getFrameSize()]);
									}
									int inByteOffset = 0;
									int frameCount = bytesRead / format.getFrameSize();	
									int outByteOffset = 0;
									int leftIndex = 0;
									int rightIndex = 1;

									FloatSampleTools.byte2float(buffer, inByteOffset, arrayList, outByteOffset, frameCount, format);
									float leftRMS= Levels.calculateRMS(arrayList.get(leftIndex));
									float rightRMS = Levels.calculateRMS(arrayList.get(rightIndex));

									float leftDBFS = 20 * (float)Math.log10(leftRMS); 
									float rightDBFS = 20 * (float)Math.log10(rightRMS);

									float leftDBSPL = leftDBFS + referenceSPL;
									float rightDBSPL = rightDBFS + referenceSPL;

									long currentTime = System.currentTimeMillis();
									if (currentTime - lastPrintTime >= 1000) {  // Check if 1000 ms (1 second) has passed
										System.out.println("Java"
											+ " Left: " 
											+ String.format("%7.2f", leftDBFS) + " dBFS " 
											+ String.format("%7.2f", leftDBSPL) + " dBFS " 
											+ String.format("%7.2f", leftDBSPL)
											+ " Right:" 
											+ String.format("%7.2f", rightDBFS) + " dBFS " 
											+ String.format("%7.2f", rightDBSPL) + " dBSPL"
										);

										lastPrintTime = currentTime;
									}
								}
            }
        } catch (LineUnavailableException e) {
            e.printStackTrace();
        }
    }

}