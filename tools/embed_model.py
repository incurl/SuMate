#!/usr/bin/env python3
"""Embedding pipeline for the SuMate vectorisation experiment."""
import json, re, sys
import numpy as np
import onnxruntime as ort

CJK = re.compile(r'[㐀-䶿一-鿿]')

class BertEmbedder:
    def __init__(self, d='/tmp/embed'):
        self.vocab = {}
        for i, line in enumerate(open(f'{d}/vocab.txt', encoding='utf-8').read().split('\n')):
            if line:
                self.vocab[line] = i
        self.unk = self.vocab.get('[UNK]', 0)
        self.cls = self.vocab.get('[CLS]', 101)
        self.sep = self.vocab.get('[SEP]', 102)
        self.pad = self.vocab.get('[PAD]', 0)
        self.sess = ort.InferenceSession(f'{d}/model_O4.onnx',
                                        providers=['CPUExecutionProvider'])
        self.dim = self.sess.get_outputs()[0].shape[-1]

    def tokenize(self, text, maxlen=128):
        toks = [self.cls]
        for ch in text:
            if ch in ('\n', '\r', '\t'):
                continue
            if not CJK.match(ch) and not ch.isalnum():
                continue
            toks.append(self.vocab.get(ch, self.unk))
        toks.append(self.sep)
        if len(toks) > maxlen:
            toks = toks[:maxlen-1] + [self.sep]
        return toks

    def embed(self, texts):
        allids = [self.tokenize(t) for t in texts]
        L = max(len(x) for x in allids)
        n = len(allids)
        ids = np.zeros((n, L), dtype=np.int64)
        mask = np.zeros((n, L), dtype=np.int64)
        ttype = np.zeros((n, L), dtype=np.int64)
        for i, t in enumerate(allids):
            ids[i, :len(t)] = t
            mask[i, :len(t)] = 1
        out = self.sess.run(None, {
            'input_ids': ids, 'attention_mask': mask, 'token_type_ids': ttype})
        last = out[0]
        # mean pooling over the real tokens
        m = mask[..., None].astype(np.float32)
        emb = (last * m).sum(1) / np.clip(m.sum(1), 1e-9, None)
        return emb / (np.linalg.norm(emb, axis=1, keepdims=True) + 1e-9)

if __name__ == '__main__':
    e = BertEmbedder()
    print('dim', e.dim)
    v = e.embed(['我想一个人待着', '明月几时有'])
    print('sim', float(v[0] @ v[1]))
