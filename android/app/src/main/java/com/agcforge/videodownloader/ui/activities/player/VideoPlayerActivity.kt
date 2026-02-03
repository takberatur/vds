package com.agcforge.videodownloader.ui.activities.player

import android.content.Context
import android.content.Intent
import android.content.pm.ActivityInfo
import android.net.Uri
import android.os.Bundle
import android.view.View
import android.view.WindowManager
import androidx.activity.OnBackPressedCallback
import androidx.appcompat.app.AppCompatActivity
import androidx.core.view.WindowCompat
import androidx.core.view.WindowInsetsCompat
import androidx.core.view.WindowInsetsControllerCompat
import androidx.media3.common.MediaItem
import androidx.media3.common.PlaybackException
import androidx.media3.common.Player
import androidx.media3.common.util.UnstableApi
import androidx.media3.exoplayer.ExoPlayer
import androidx.media3.exoplayer.trackselection.DefaultTrackSelector
import com.agcforge.videodownloader.databinding.ActivityVideoPlayerBinding
import com.agcforge.videodownloader.utils.showToast
import com.google.android.material.dialog.MaterialAlertDialogBuilder
import androidx.core.net.toUri
import androidx.core.view.GravityCompat
import com.agcforge.videodownloader.R


@UnstableApi
class VideoPlayerActivity : AppCompatActivity() {
    private lateinit var binding: ActivityVideoPlayerBinding
    private var player: ExoPlayer? = null
    private var trackSelector: DefaultTrackSelector? = null

    private var videoUri: Uri? = null
    private var videoTitle: String = ""
    private var currentPlaybackPosition: Long = 0
    private var isLocked = false
    private var isFullscreen = false

    companion object {
        private const val KEY_VIDEO_URI = "video_uri"
        private const val KEY_VIDEO_TITLE = "video_title"
        private const val KEY_PLAYBACK_POSITION = "playback_position"

        fun start(context: Context, videoUri: String, title: String) {
            val intent = Intent(context, VideoPlayerActivity::class.java).apply {
                putExtra(KEY_VIDEO_URI, videoUri)
                putExtra(KEY_VIDEO_TITLE, title)
            }
            context.startActivity(intent)
        }
    }

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        binding = ActivityVideoPlayerBinding.inflate(layoutInflater)
        setContentView(binding.root)

        setupBackPressedCallback()

        // Get video info from intent
        videoUri = intent.getStringExtra(KEY_VIDEO_URI)?.toUri()
        videoTitle = intent.getStringExtra(KEY_VIDEO_TITLE) ?: "Video Player"

        if (savedInstanceState != null) {
            currentPlaybackPosition = savedInstanceState.getLong(KEY_PLAYBACK_POSITION, 0)
        }

        setupUI()
        setupPlayer()
        setupListeners()
    }

    private fun setupUI() {
        // Set video title
        binding.tvVideoTitle.text = videoTitle

        // Hide system UI for immersive mode
        hideSystemUI()

        // Keep screen on
        window.addFlags(WindowManager.LayoutParams.FLAG_KEEP_SCREEN_ON)
    }

    private fun setupPlayer() {
        // Create track selector for quality selection
        trackSelector = DefaultTrackSelector(this).apply {
            setParameters(buildUponParameters().setMaxVideoSizeSd())
        }

        // Initialize ExoPlayer
        player = ExoPlayer.Builder(this)
            .setTrackSelector(trackSelector!!)
            .build()
            .also { exoPlayer ->
                binding.playerView.player = exoPlayer

                // Set media item
                videoUri?.let { uri ->
                    val mediaItem = MediaItem.fromUri(uri)
                    exoPlayer.setMediaItem(mediaItem)
                    exoPlayer.prepare()

                    // Seek to saved position
                    if (currentPlaybackPosition > 0) {
                        exoPlayer.seekTo(currentPlaybackPosition)
                    }

                    exoPlayer.playWhenReady = true
                }

                // Add player listener
                exoPlayer.addListener(object : Player.Listener {
                    override fun onPlaybackStateChanged(playbackState: Int) {
                        when (playbackState) {
                            Player.STATE_BUFFERING -> {
                                binding.progressLoading.visibility = View.VISIBLE
                            }
                            Player.STATE_READY -> {
                                binding.progressLoading.visibility = View.GONE
                                binding.llError.visibility = View.GONE
                            }
                            Player.STATE_ENDED -> {
                                // Video ended
                                exoPlayer.seekTo(0)
                                exoPlayer.pause()
                            }

                            Player.STATE_IDLE -> {
                                TODO()
                            }
                        }
                    }

                    override fun onPlayerError(error: PlaybackException) {
                        showError("Failed to play video: ${error.message}")
                    }
                })
            }
    }

    private fun setupListeners() {
        // Back button
        binding.btnBack.setOnClickListener {
            finish()
        }

        // Lock button
        binding.btnLock.setOnClickListener {
            toggleLock()
        }

        // Retry button
        binding.btnRetry.setOnClickListener {
            binding.llError.visibility = View.GONE
            setupPlayer()
        }

        // Quality button (from custom controls)
        binding.playerView.findViewById<View>(com.agcforge.videodownloader.R.id.btnQuality)?.setOnClickListener {
            showQualityDialog()
        }

        // Speed button
        binding.playerView.findViewById<View>(com.agcforge.videodownloader.R.id.btnSpeed)?.setOnClickListener {
            showSpeedDialog()
        }

        // Fullscreen button
        binding.playerView.findViewById<View>(com.agcforge.videodownloader.R.id.btnFullscreen)?.setOnClickListener {
            toggleFullscreen()
        }

        // Subtitle button
        binding.playerView.findViewById<View>(com.agcforge.videodownloader.R.id.btnSubtitle)?.setOnClickListener {
            showToast("Subtitle feature coming soon")
        }
    }

    private fun toggleLock() {
        isLocked = !isLocked

        if (isLocked) {
            // Hide all controls except lock button
            binding.playerView.hideController()
            binding.btnBack.visibility = View.GONE
            binding.tvVideoTitle.visibility = View.GONE
            binding.btnLock.setImageResource(com.agcforge.videodownloader.R.drawable.ic_lock_white)
            showToast("Controls locked")
        } else {
            // Show controls
            binding.playerView.showController()
            binding.btnBack.visibility = View.VISIBLE
            binding.tvVideoTitle.visibility = View.VISIBLE
            binding.btnLock.setImageResource(com.agcforge.videodownloader.R.drawable.ic_lock_open_white)
            showToast("Controls unlocked")
        }
    }

    private fun toggleFullscreen() {
        isFullscreen = !isFullscreen

        if (isFullscreen) {
            // Enter fullscreen (landscape)
            requestedOrientation = ActivityInfo.SCREEN_ORIENTATION_SENSOR_LANDSCAPE
        } else {
            // Exit fullscreen (portrait)
            requestedOrientation = ActivityInfo.SCREEN_ORIENTATION_PORTRAIT
        }
    }

    private fun showQualityDialog() {
        val qualities = arrayOf("Auto", "1080p", "720p", "480p", "360p")

        MaterialAlertDialogBuilder(this)
            .setTitle(getString(R.string.video_quality))
            .setItems(qualities) { dialog, which ->
                showToast("Quality set to ${qualities[which]}")
                dialog.dismiss()
            }
            .show()
    }

    private fun showSpeedDialog() {
        val speeds = arrayOf("0.25x", "0.5x", "0.75x", "1.0x", "1.25x", "1.5x", "1.75x", "2.0x")
        val speedValues = floatArrayOf(0.25f, 0.5f, 0.75f, 1.0f, 1.25f, 1.5f, 1.75f, 2.0f)

        MaterialAlertDialogBuilder(this)
            .setTitle(getString(R.string.playback_speed))
            .setItems(speeds) { dialog, which ->
                player?.setPlaybackSpeed(speedValues[which])
                binding.playerView.findViewById<android.widget.TextView>(
                    com.agcforge.videodownloader.R.id.btnSpeed
                )?.text = speeds[which]
                showToast("Speed set to ${speeds[which]}")
                dialog.dismiss()
            }
            .show()
    }

    private fun showError(message: String) {
        binding.llError.visibility = View.VISIBLE
        binding.tvErrorMessage.text = message
        binding.progressLoading.visibility = View.GONE
    }

    private fun hideSystemUI() {
        WindowCompat.setDecorFitsSystemWindows(window, false)
        WindowInsetsControllerCompat(window, binding.root).let { controller ->
            controller.hide(WindowInsetsCompat.Type.systemBars())
            controller.systemBarsBehavior = WindowInsetsControllerCompat.BEHAVIOR_SHOW_TRANSIENT_BARS_BY_SWIPE
        }
    }

    override fun onSaveInstanceState(outState: Bundle) {
        super.onSaveInstanceState(outState)
        outState.putLong(KEY_PLAYBACK_POSITION, player?.currentPosition ?: 0)
    }

    override fun onPause() {
        super.onPause()
        player?.pause()
    }

    override fun onStop() {
        super.onStop()
        currentPlaybackPosition = player?.currentPosition ?: 0
    }

    override fun onDestroy() {
        super.onDestroy()
        releasePlayer()
    }

    private fun releasePlayer() {
        player?.let {
            currentPlaybackPosition = it.currentPosition
            it.release()
        }
        player = null
        trackSelector = null
    }

    private fun setupBackPressedCallback() {
        val callback = object : OnBackPressedCallback(true) {
            override fun handleOnBackPressed() {
                if (isFullscreen) {
                    toggleFullscreen()
                } else {
                    onBackPressedDispatcher.onBackPressed()
                }
            }
        }
        onBackPressedDispatcher.addCallback(this, callback)
    }
}