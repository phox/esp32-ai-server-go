import sys
import json
import soundfile as sf
import numpy as np
import onnxruntime as ort
import argparse
import os

def run_vad(wav_path, model_path, threshold=0.5, sample_rate=16000):
    # 读取音频
    audio, sr = sf.read(wav_path)
    if sr != sample_rate:
        print(json.dumps({"error": f"Sample rate mismatch: {sr} != {sample_rate}"}))
        sys.exit(1)
    if audio.ndim > 1:
        audio = audio.mean(axis=1)  # 转为单通道
    # 归一化
    if np.abs(audio).max() > 1.0:
        audio = audio / np.abs(audio).max()
    # 加载ONNX模型
    if not os.path.exists(model_path):
        print(json.dumps({"error": f"Model not found: {model_path}"}))
        sys.exit(1)
    ort_sess = ort.InferenceSession(model_path)
    # 按silero官方推理流程
    # https://github.com/snakers4/silero-vad/blob/master/utils_vad.py
    # 输入shape: [batch, samples], 需float32
    input_audio = audio.astype(np.float32)[None, :]
    # 推理
    outs = ort_sess.run(None, {ort_sess.get_inputs()[0].name: input_audio})
    speech_probs = outs[0][0]  # [samples]
    # VAD后处理，找出大于阈值的区间
    segments = []
    in_speech = False
    seg_start = 0
    for i, prob in enumerate(speech_probs):
        if not in_speech and prob > threshold:
            in_speech = True
            seg_start = i
        elif in_speech and prob <= threshold:
            in_speech = False
            segments.append([seg_start, i])
    if in_speech:
        segments.append([seg_start, len(speech_probs)])
    print(json.dumps(segments))

if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('--input', required=True)
    parser.add_argument('--model', default='models/silero_vad.onnx')
    parser.add_argument('--threshold', type=float, default=0.5)
    parser.add_argument('--sample-rate', type=int, default=16000)
    args = parser.parse_args()
    run_vad(args.input, args.model, args.threshold, args.sample_rate) 