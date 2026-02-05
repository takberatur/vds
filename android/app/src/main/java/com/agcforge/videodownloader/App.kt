package com.agcforge.videodownloader

import android.annotation.SuppressLint
import android.app.Application
import android.content.Context
import androidx.lifecycle.DefaultLifecycleObserver
import androidx.lifecycle.LifecycleOwner
import androidx.lifecycle.ProcessLifecycleOwner
import com.agcforge.videodownloader.helper.AppOpenAdManager
import com.agcforge.videodownloader.ui.activities.BaseActivity
import com.agcforge.videodownloader.utils.DownloadManagerCleaner
import com.onesignal.OneSignal
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob

import kotlinx.coroutines.launch



class App : Application(), DefaultLifecycleObserver {

    companion object {
        @SuppressLint("StaticFieldLeak")
        @Volatile
        private var instance: App? = null

        fun getInstance(): App {
            return instance ?: synchronized(this) {
                instance ?: throw IllegalStateException("App not initialized")
            }
        }

        fun getContext(): Context {
            return instance?.applicationContext ?: throw IllegalStateException("Context not initialized")
        }
    }

    private val applicationScope = CoroutineScope(SupervisorJob() + Dispatchers.Main)

    override fun attachBaseContext(base: Context?) {
        super.attachBaseContext(base)
        instance = this
    }

    override fun onCreate(owner: LifecycleOwner) {
        super.onStart(owner)
        instance = this

        ProcessLifecycleOwner.get().lifecycle.addObserver(this)

        DownloadManagerCleaner.clearFailedDownloads(this)
        AppOpenAdManager.loadAd(this)

    }

    override fun onStop(owner: LifecycleOwner) {
        super.onStop(owner)
        AppOpenAdManager.onAppEnteredBackground()
    }

    override fun onStart(owner: LifecycleOwner) {
        super.onStart(owner)
        BaseActivity.CurrentActivityHolder.activity?.let {
            AppOpenAdManager.showAdIfAvailable(it)
        }
    }
}
