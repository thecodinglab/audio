#include <pipewire/pipewire.h>
#include <spa/param/audio/format-utils.h>

struct data {
  int sample_rate;
  int channels;

  struct pw_main_loop *loop;
  struct pw_stream *stream;
  void *userdata;
};

void audio_sample(void *buf, size_t size, void *userdata);

static void on_process(void *userdata) {
  struct data *data = userdata;

  struct pw_buffer *b = pw_stream_dequeue_buffer(data->stream);
  if (b == NULL)
    return;

  struct spa_buffer *buf = b->buffer;

  int16_t *dst = buf->datas[0].data;
  if (dst == NULL)
    return;

  int stride = sizeof(int16_t) * data->channels;
  int n_frames = buf->datas[0].maxsize / stride;
  if (b->requested)
    n_frames = SPA_MIN(b->requested, n_frames);

  audio_sample(dst, n_frames * stride, data->userdata);

  buf->datas[0].chunk->offset = 0;
  buf->datas[0].chunk->stride = stride;
  buf->datas[0].chunk->size = n_frames * stride;

  pw_stream_queue_buffer(data->stream, b);
}

static const struct pw_stream_events stream_events = {
    PW_VERSION_STREAM_EVENTS,
    .process = on_process,
};

struct data *audio_setup(const char *name, int sampleRate, int channels,
                         void *userdata) {
  struct data *data = malloc(sizeof(struct data));
  data->sample_rate = sampleRate;
  data->channels = channels;
  data->userdata = userdata;

  pw_init(NULL, NULL);
  data->loop = pw_main_loop_new(NULL);

  uint8_t buffer[1024];
  struct spa_pod_builder builder = SPA_POD_BUILDER_INIT(buffer, sizeof(buffer));

  const struct spa_pod *params[1];

  params[0] = spa_format_audio_raw_build(
      &builder, SPA_PARAM_EnumFormat,
      &SPA_AUDIO_INFO_RAW_INIT(.format = SPA_AUDIO_FORMAT_S16,
                               .channels = channels, .rate = sampleRate));

  data->stream = pw_stream_new_simple(
      pw_main_loop_get_loop(data->loop), name,
      pw_properties_new(PW_KEY_MEDIA_TYPE, "Audio",        //
                        PW_KEY_MEDIA_CATEGORY, "Playback", //
                        PW_KEY_MEDIA_ROLE, "Music",        //
                        NULL),
      &stream_events, data);

  pw_stream_connect(data->stream, PW_DIRECTION_OUTPUT, PW_ID_ANY,
                    PW_STREAM_FLAG_AUTOCONNECT | PW_STREAM_FLAG_MAP_BUFFERS |
                        PW_STREAM_FLAG_RT_PROCESS,
                    params, 1);

  return data;
}

void audio_run(struct data *data) { pw_main_loop_run(data->loop); }

void audio_quit(struct data *data) { pw_main_loop_quit(data->loop); }

void audio_close(struct data *data) {
  pw_stream_destroy(data->stream);
  pw_main_loop_destroy(data->loop);
  free(data);
}
